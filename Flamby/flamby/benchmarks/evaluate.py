import io
import copy

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
@app.route('/evaluate', methods=['POST'])
def main():

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
    client = request.get_json()
    client_id = range(6)
    local_updates = list()
    total_number_of_samples = 0

    for id in client_id:
        print("id:", id)
        updates = load_update(client[id]["update_hash"])
        #list转numpy
        for arr in updates:
            arr = np.array(arr)
        local_updates.append({"updates": updates, "n_samples": client[id]["n_samples"]})
        total_number_of_samples += client[id]["n_samples"]

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
    initial_model = Baseline()
    global_model_previous = initial_model.load_state_dict(load_model(client["global_model_hash"]))
    global_model = global_model_previous.deepcopy()
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
        global_model_test = global_model_previous.deepcopy()
        local_previous_state_1 = [
            param.cpu().detach().clone().numpy() for param in global_model_test.parameters()
        ]

        for old_param, new_param in zip(global_model_test.parameters(), aggregated_delta_weights):
            old_param.data += torch.from_numpy(new_param).to(old_param.device)

        pooled_perf_dict = evaluate_model_on_local_and_pooled_tests(global_model_test.model, test_dls,
                                                                    test_pooled, metric, evaluate_func,
                                                                    local=False)
        performance[f"except_client_{idx_client}"] = pooled_perf_dict['client_test_0']
        print(f"except_client_{idx_client}: ", performance[f"except_client_{idx_client}"])

        for p_new, p_old in zip(global_model_test.model.parameters(), local_previous_state_1):
            p_new.data = torch.from_numpy(p_old).to(p_new.device)
        del local_previous_state_1

    best_except = []
    contribute_list = []
    num_clients = len(client_id)
    for i in range(num_clients):
        contribute = performance[f"avg_gloabl_model"] - performance[f"except_client_{i}"]
        if contribute < -0.009:
            best_except.append(i)
            contribute = 0
        elif contribute < 0:
            contribute = contribute * (-0.08)
        contribute_list.append(contribute)
    contribute_sum = sum(contribute_list)
    contribute_list.append(contribute_sum)
    print("contribute:", contribute_list)
    #self.writer1.writerow(contribute_list)

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
                    0.3 * local_updates[idx_client]["n_samples"] / join_counts + 0.7 * contribute_list[
                        idx_client] / contribute_sum))

    #上传基于贡献聚合的模型
    for old_param, new_param in zip(global_model_previous.parameters(), aggregated_delta_weights):
        old_param.data += torch.from_numpy(new_param).to(old_param.device)

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

if __name__ == "__main__":
    app.run(port=6008)
