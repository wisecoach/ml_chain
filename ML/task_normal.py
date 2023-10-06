import copy
import agent
import numpy
import trainers
import dataset
import model_2
import torch

import verifiers

device = torch.device("cuda" if torch.cuda.is_available() else "cpu")


class Task:
    def __init__(self, genesis):
        self.iteration = 0
        self.model_structure = genesis['model_structure']
        self.dataset_train, self.dataset_test = dataset.get_dataset(self.model_structure)
        self.trainer_dict = {}
        if self.model_structure['dataset'] == 'mnist':
            self.global_model = model_2.MNIST_CNN_Net()
        else:
            self.global_model = model_2.CIFAR_CNN_Net()
        # self.global_model.load_state_dict()
        init_weights = numpy.array(genesis['init_weight']['payload']['weight_vector'])
        layers = self.global_model.reshape(init_weights)
        layer = 0
        for name, param in self.global_model.named_parameters():
            if param.requires_grad:
                param.data = layers[layer]
                layer += 1

    def train(self, req):
        trainer_id = req['trainer_id']
        weight = req['global_model']['weight_vector']
        layers = self.global_model.reshape(weight)
        layer = 0
        for name, param in self.global_model.named_parameters():
            if param.requires_grad:
                param.data = layers[layer]
                layer += 1
        if not(trainer_id in self.trainer_dict):
            trainer = trainers.Trainer(self.model_structure, self.global_model, self.dataset_train, self.dataset_test)
            self.trainer_dict[trainer_id] = trainer
        trainer = self.trainer_dict[trainer_id]
        w, acc, loss, data_num = trainer.local_update(device, self.iteration)
        flat_w = []
        for value in w.values():
            if device.type == 'cuda':
                value = value.cpu()
            flat_w = numpy.append(flat_w, value.reshape(-1).data.numpy())
        train_resp = {
            'local_model': {
                'weight_vector': flat_w.tolist(),
                'acc': acc,
                'loss': loss,
                'data_num': data_num
            },
        }
        return train_resp

    def validate(self, req):
        weight = req['model']['weight_vector']
        test_model = copy.deepcopy(self.global_model)
        layers = test_model.reshape(weight)
        layer = 0
        for name, param in test_model.named_parameters():
            if param.requires_grad:
                param.data = layers[layer]
                layer += 1
        verifier = verifiers.Verifier(self.model_structure, self.dataset_test)
        test_loss, acc = verifier.test(test_model, device)
        resp = {
            'loss': {
                'loss': test_loss,
                'acc': acc,
            }
        }
        return resp

    def aggregate(self, req):
        local_models = req['local_models']
        return self.aggregate_impl1(local_models)

    # 加权平均聚合
    def aggregate_impl1(self, local_models):
        data_num_list = []
        data_sum = 0
        w = []
        for local_model in local_models:
            # 取出本地模型的数据量
            data_num_list.append(local_model['data_num'])
            data_sum += local_model['data_num']
            # 取得局部模型参数
            params_dict = {}
            weight = local_model['weight_vector']
            layers = self.global_model.reshape(weight)
            layer = 0
            for name, param in self.global_model.named_parameters():
                if param.requires_grad:
                    params_dict[name] = layers[layer]
                    layer += 1
            w.append(params_dict)
        # 计算出本地模型在聚合时的权重以及Si
        agg_weights = [data_num / data_sum for data_num in data_num_list]
        # 聚合模型
        new_global_weight = aggregate_model(w, agg_weights)
        # 导入聚合好的模型
        verifier = agent.Agent(self.model_structure, self.dataset_test, new_global_weight)
        total_acc = verifier.evaluate()

        flat_new_global_weight = []
        for value in new_global_weight.values():
            flat_new_global_weight = numpy.append(flat_new_global_weight, value.reshape(-1).data.numpy())
        return {
            'global_model': {
                'weight_vector': flat_new_global_weight.tolist(),
                'total_acc': total_acc,
            }
        }

    # 直接平均聚合
    def aggregate_impl2(self, local_models):
        w = []
        for local_model in local_models:
            # 取得局部模型参数
            params_dict = {}
            weight = local_model['weight_vector']
            layers = self.global_model.reshape(weight)
            layer = 0
            for name, param in self.global_model.named_parameters():
                if param.requires_grad:
                    params_dict[name] = layers[layer]
                    layer += 1
            w.append(params_dict)
        # 计算出本地模型在聚合时的权重以及Si
        agg_weights = [1 / len(local_models) for i in range(len(local_models))]
        # 聚合模型
        new_global_weight = aggregate_model(w, agg_weights)
        # 导入聚合好的模型
        verifier = agent.Agent(self.model_structure, self.dataset_test, new_global_weight)
        total_acc = verifier.evaluate()

        flat_new_global_weight = []
        for value in new_global_weight.values():
            flat_new_global_weight = numpy.append(flat_new_global_weight, value.reshape(-1).data.numpy())
        return {
            'global_model': {
                'weight_vector': flat_new_global_weight.tolist(),
                'total_acc': total_acc,
            }
        }


def aggregate_model(w, agg_weight):
    # 创建一个和第一个模型权重相同形状的全零张量
    global_model = copy.deepcopy(w[0])
    for k in global_model.keys():
        global_model[k] = torch.zeros_like(w[0][k])

    # 对每个模型的权重进行加权相加
    for i in range(len(w)):
        for k in global_model.keys():
            global_model[k] += w[i][k] * agg_weight[i]

    return global_model
