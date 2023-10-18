import copy
import io

from flask import Flask, request
from flask_cors import CORS
import torch
from flamby.strategies.utils import DataLoaderWithMemory
# 2 lines of code to change to switch to another dataset
from flamby.datasets.fed_isic2019 import (
    BATCH_SIZE,
    LR,
    Baseline,
    BaselineLoss,
    Optimizer,
)
from flamby.datasets.fed_isic2019 import FedIsic2019 as FedDataset
import numpy as np
from flask import Flask, request
from flask_cors import CORS
import torch
from flamby.benchmarks.conf import get_dataset_args
from flamby.benchmarks.benchmark_utils import init_data_loaders
from flamby.benchmarks.benchmark_utils import set_dataset_specific_config
from flamby.benchmarks.benchmark_utils import evaluate_model_on_local_and_pooled_tests
from ipfs import save_model, load_model, save_update, load_update


app = Flask(__name__)
CORS(app, supports_credentials=True)


@app.route('/train', methods=['GET'])
def train():
    cindex = request.args.get("cindex")
    model_hash = request.args.get("model_hash")
    print("cindex:", cindex)
    print("model_hash", model_hash)
    train_dataset = FedDataset(center=int(cindex), train=True, pooled=False)
    train_dataloader = torch.utils.data.DataLoader(train_dataset, batch_size=BATCH_SIZE, shuffle=True, num_workers=0)
    training_dataloader_with_memory = DataLoaderWithMemory(train_dataloader)
    size = len(training_dataloader_with_memory)
    print("size:", size)
    lossfunc = BaselineLoss()
    model = Baseline()
    model.load_state_dict(load_model(model_hash))
    if torch.cuda.is_available():
        model.cuda()
    optimizer = Optimizer(model.parameters(), lr=LR)
    
    local_previous_state = [
            param.cpu().detach().clone().numpy() for param in model.parameters()
        ]
    
    # Traditional pytorch training loop
    local_train_num = 1
    for epoch in range(0, local_train_num):
        print("本地运行轮次数：", epoch)
        for idx, (X, y) in enumerate(train_dataloader):
            X, y = X.cuda(), y.cuda()
            optimizer.zero_grad()
            outputs = model(X)
            loss = lossfunc(outputs, y)
            loss.backward()
            optimizer.step()
    local_next_state = [
            param.cpu().detach().clone().numpy() for param in model.parameters()
        ]
    
    # Recovering updates
    updates = [
        (new - old).tolist() for new, old in zip(local_next_state, local_previous_state)
    ]
    del local_next_state
    new_model_hash = save_model(model.state_dict())
        
    update_hash = save_update(updates)
    print({"local_model": {"cindex": cindex, "n_samples": size, "model_hash": new_model_hash, "update_hash": update_hash}})
    return {"local_model": {"n_samples": size, "model_hash": new_model_hash, "update_hash": update_hash}}


@app.route('/aggregate', methods=['POST'])
def aggregate():
    evaluate_func, batch_size_test, compute_ensemble_perf = set_dataset_specific_config(
        "fed_isic2019", compute_ensemble_perf=False
    )

    (
        FedDataset,
        [
            BATCH_SIZE,
            LR,
            NUM_CLIENTS,
            NUM_EPOCHS_POOLED,
            Baseline,
            BaselineLoss,
            Optimizer,
            get_nb_max_rounds,
            metric,
            collate_fn,
        ],
    ) = get_dataset_args("fed_isic2019")

    training_dls, test_dls = init_data_loaders(
        dataset=FedDataset,
        pooled=False,
        batch_size=BATCH_SIZE,
        num_workers=0,
        num_clients=NUM_CLIENTS,
        batch_size_test=batch_size_test,
        collate_fn=collate_fn,
    )
    train_pooled, test_pooled = init_data_loaders(
        dataset=FedDataset,
        pooled=True,
        batch_size=BATCH_SIZE,
        num_workers=0,
        batch_size_test=batch_size_test,
        collate_fn=collate_fn,
    )

    #计算各节点的updates
    req = request.get_json()
    local_models = req['local_models']
    last_global_model = req['last_global_model']
    client_id = [str(i) for i in range(6)]
    local_updates = list()
    total_number_of_samples = 0

    for id in client_id:
        print("id:", id)
        updates = load_update(local_models[id]["update_hash"])
        #list转numpy
        nd_updates = []
        for arr in updates:
            nd_updates.append(np.array(arr))
        local_updates.append({"updates": nd_updates, "n_samples": local_models[id]["n_samples"]})
        total_number_of_samples += local_models[id]["n_samples"]

    #fedavg策略计算梯度
    aggregated_delta_weights = [
        None for _ in range(len(local_updates[0]["updates"]))
    ]
    for idx_weight in range(len(local_updates[0]["updates"])):
        aggregated_delta_weights[idx_weight] = sum(
            [
                local_updates[idx_client]["updates"][idx_weight] * local_updates[idx_client]["n_samples"]
                for idx_client in range(6)
            ]
        )
        aggregated_delta_weights[idx_weight] /= float(total_number_of_samples)

    #上一轮的全局模型
    global_model_previous = Baseline()
    global_model_previous.load_state_dict(load_model(last_global_model['model_hash']))
    global_model = copy.deepcopy(global_model_previous)
    local_previous_state = [
            param.cpu().detach().clone().numpy() for param in global_model.parameters()
        ]

    #更新fedavg下全局模型梯度
    for old_param, new_param in zip(global_model.parameters(), aggregated_delta_weights):
        old_param.data += torch.from_numpy(new_param).to(old_param.device)

    performance = {}

    # fedavg下的全局模型精度
    pooled_perf_dict = evaluate_model_on_local_and_pooled_tests(global_model, test_dls, test_pooled,
                                                                  metric, evaluate_func, local=False)
    performance["avg_gloabl_model"] = pooled_perf_dict['client_test_0']
    print("avg_gloabl_model", performance["avg_gloabl_model"])

    # Reset local model
    for p_new, p_old in zip(global_model.parameters(), local_previous_state):
        p_new.data = torch.from_numpy(p_old).to(p_new.device)
    del local_previous_state

    # 测量每个节点的贡献
    for idx_client in range(6):
        aggregated_delta_weights = [
            None for _ in range(len(local_updates[0]["updates"]))
        ]
        for idx_weight in range(len(local_updates[0]["updates"])):
            aggregated_delta_weights[idx_weight] = sum(
                [
                    local_updates[idx]["updates"][idx_weight] * local_updates[idx]["n_samples"]
                    for idx in range(6)
                ]
            )
            aggregated_delta_weights[idx_weight] -= local_updates[idx_client]["updates"][idx_weight] * local_updates[idx_client]["n_samples"]
            aggregated_delta_weights[idx_weight] /= float(
                total_number_of_samples - local_updates[idx_client]["n_samples"])

        # 初始化全局模型
        global_model_test = copy.deepcopy(global_model_previous)
        local_previous_state_1 = [
            param.cpu().detach().clone().numpy() for param in global_model_test.parameters()
        ]

        for old_param, new_param in zip(global_model_test.parameters(), aggregated_delta_weights):
            old_param.data += torch.from_numpy(new_param).to(old_param.device)

        pooled_perf_dict = evaluate_model_on_local_and_pooled_tests(global_model_test, test_dls,
                                                                    test_pooled, metric, evaluate_func,
                                                                    local=False)
        performance[f"except_client_{idx_client}"] = pooled_perf_dict['client_test_0']
        print(f"except_client_{idx_client}: ", performance[f"except_client_{idx_client}"])

        for p_new, p_old in zip(global_model_test.parameters(), local_previous_state_1):
            p_new.data = torch.from_numpy(p_old).to(p_new.device)
        del local_previous_state_1

    best_except = []
    contributes = {}
    contribute_sum = 0.0
    num_clients = len(client_id)
    for i in range(num_clients):
        contribute = performance[f"avg_gloabl_model"] - performance[f"except_client_{i}"]
        if contribute < -0.009:
            best_except.append(i)
            contribute = 0
        elif contribute < 0:
            contribute = contribute * (-0.08)
        contributes[str(i)] = contribute
        contribute_sum += contribute

    # 根据贡献聚合全局模型
    aggregated_delta_weights = [
        0.0 for _ in range(len(local_updates[0]["updates"]))
    ]
    no_join = 0
    print("best_except", best_except)
    for id in best_except:
        print("id:", id)
        no_join += local_updates[id]["n_samples"]
    join_counts = total_number_of_samples - no_join
    for idx_weight in range(len(local_updates[0]["updates"])):
        for idx_client in range(6):
            if idx_client not in best_except:
                aggregated_delta_weights[idx_weight] += (local_updates[idx_client]["updates"][idx_weight] * float(
                    0.3 * local_updates[idx_client]["n_samples"] / join_counts + 0.7 * contributes[str(idx_client)]
                    / contribute_sum))

    #上传基于贡献聚合的模型
    for old_param, new_param in zip(global_model_previous.parameters(), aggregated_delta_weights):
        old_param.data += torch.from_numpy(new_param).to(old_param.device)

    model_hash = save_model(global_model_previous.state_dict())

    # 计算贡献全局模型的精度
    (
        perf_dict,
        pooled_perf_dict,
        y_true_dict,
        y_pred_dict,
        y_true_pooled_dict,
        y_pred_pooled_dict,
    ) = evaluate_model_on_local_and_pooled_tests(global_model_previous, test_dls, test_pooled,
                                                   metric, evaluate_func, local=True)
    # print("Global Model Per-center performance:")
    print(perf_dict)
    print("Global Model Performance on pooled test set: ", pooled_perf_dict['client_test_0'])
    return {"global_model": {"model_hash": model_hash}, "contributes": contributes}


if __name__ == "__main__":
    app.run(port=6006, threaded=False)
