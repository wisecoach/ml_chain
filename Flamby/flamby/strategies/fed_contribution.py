import time
import copy
from typing import List
import csv
import torch
from tqdm import tqdm
from flamby.utils import evaluate_model_on_tests, evaluate_model_on_tests_in_specify_index
from flamby.benchmarks.benchmark_utils import evaluate_model_on_local_and_pooled_tests
from flamby.strategies.utils import DataLoaderWithMemory, _Model

class FedContribution:
    """Federated Averaging Strategy class.

    The Federated Averaging strategy is the most simple centralized FL strategy.
    Each client first trains his version of a global model locally on its data,
    the states of the model of each client are then weighted-averaged and returned
    to each client for further training.

    References
    ----------
    - https://arxiv.org/abs/1602.05629

    Parameters
    ----------
    training_dataloaders : List
        The list of training dataloaders from multiple training centers.
    model : torch.nn.Module
        An initialized torch model.
    loss : torch.nn.modules.loss._Loss
        The loss to minimize between the predictions of the model and the
        ground truth.
    optimizer_class : torch.optim.Optimizer
        The class of the torch model optimizer to use at each step.
    learning_rate : float
        The learning rate to be given to the optimizer_class.
    num_updates : int
        The number of updates to do on each client at each round.
    nrounds : int
        The number of communication rounds to do.
    dp_target_epsilon: float
        The target epsilon for (epsilon, delta)-differential
        private guarantee. Defaults to None.
    dp_target_delta: float
        The target delta for (epsilon, delta)-differential private
        guarantee. Defaults to None.
    dp_max_grad_norm: float
        The maximum L2 norm of per-sample gradients; used to enforce
        differential privacy. Defaults to None.
    log: bool, optional
        Whether or not to store logs in tensorboard. Defaults to False.
    log_period: int, optional
        If log is True then log the loss every log_period batch updates.
        Defauts to 100.
    bits_counting_function : Union[callable, None], optional
        A function making sure exchanges respect the rules, this function
        can be obtained by decorating check_exchange_compliance in
        flamby.utils. Should have the signature List[Tensor] -> int.
        Defaults to None.
    logdir: str, optional
        Where logs are stored. Defaults to ./runs.
    log_basename: str, optional
        The basename of the created log_file. Defaults to fed_avg.
    """
    
    #test_global_pooled: torch.utils.data.dataloader,

    def __init__(
        self,
        training_dataloaders: List,
        test_dls: List,
        test_pooled:torch.utils.data.DataLoader,
        metric: callable, 
        evaluate_func: callable,
        model: torch.nn.Module,
        loss: torch.nn.modules.loss._Loss,
        optimizer_class: torch.optim.Optimizer,
        learning_rate: float,
        num_updates: int,
        nrounds: int,
        dp_target_epsilon: float = None,
        dp_target_delta: float = None,
        dp_max_grad_norm: float = None,
        log: bool = False,
        log_period: int = 100,
        bits_counting_function: callable = None,
        logdir: str = "./runs",
        log_basename: str = "fed_contribution",
        seed=None,
    ):
        """
        Cf class docstring
        """
        self.model = model
        self.optimizer_class =optimizer_class
        self.loss =loss
        self.learning_rate = learning_rate
        
        self.test_dls = test_dls
        self.test_pooled = test_pooled
        self.metric = metric
        self.evaluate_func = evaluate_func
        
        self._seed = seed if seed is not None else int(time.time())
        print("seed:",self._seed)
        self.training_dataloaders_with_memory = [
            DataLoaderWithMemory(e) for e in training_dataloaders
        ]
        self.training_sizes = [len(e) for e in self.training_dataloaders_with_memory]
        self.total_number_of_samples = sum(self.training_sizes)

        self.dp_target_epsilon = dp_target_epsilon
        self.dp_target_delta = dp_target_delta
        self.dp_max_grad_norm = dp_max_grad_norm

        self.log = log
        self.log_period = log_period
        self.log_basename = log_basename
        self.logdir = logdir
        
        self.models_list = [
            _Model(
                model=model,
                optimizer_class=optimizer_class,
                lr=learning_rate,
                train_dl=_train_dl,
                dp_target_epsilon=self.dp_target_epsilon,
                dp_target_delta=self.dp_target_delta,
                dp_max_grad_norm=self.dp_max_grad_norm,
                loss=loss,
                nrounds=nrounds,
                log=self.log,
                client_id=i,
                log_period=self.log_period,
                log_basename=self.log_basename,
                logdir=self.logdir,
                seed=self._seed,
            )
            for i, _train_dl in enumerate(training_dataloaders)
        ]
        
        #修改了nrounds
        #self.nrounds = nrounds
        self.nrounds = 60
        self.num_updates = num_updates

        self.num_clients = len(self.training_sizes)
        print("num_clients:",self.num_clients)
        self.bits_counting_function = bits_counting_function

    def _local_optimization(self, _model: _Model, dataloader_with_memory):
        """Carry out the local optimization step.

        Parameters
        ----------
        _model: _Model
            The model on the local device used by the optimization step.
        dataloader_with_memory : dataloaderwithmemory
            A dataloader that can be called infinitely using its get_samples()
            method.
        """
        _model._local_train(dataloader_with_memory, self.num_updates)

    def perform_round(self):
        """Does a single federated averaging round. The following steps will be
        performed:

        - each model will be trained locally for num_updates batches.
        - the parameter updates will be collected and averaged. Averages will be
          weighted by the number of samples in each client
        - the averaged updates willl be used to update the local model
        """
        local_updates = list()
        performance = {}
        
        for _model, dataloader_with_memory, size in zip(
            self.models_list, self.training_dataloaders_with_memory, self.training_sizes
        ):
            # Local Optimization
            _local_previous_state = _model._get_current_params()
            self._local_optimization(_model, dataloader_with_memory)
            _local_next_state = _model._get_current_params()
            
            # Recovering updates
            updates = [
                new - old for new, old in zip(_local_next_state, _local_previous_state)
            ]
            del _local_next_state

            # Reset local model
            for p_new, p_old in zip(_model.model.parameters(), _local_previous_state):
                p_new.data = torch.from_numpy(p_old).to(p_new.device)
            del _local_previous_state

            if self.bits_counting_function is not None:
                self.bits_counting_function(updates)

            local_updates.append({"updates": updates, "n_samples": size})

            
        # Aggregation step
        aggregated_delta_weights = [
            None for _ in range(len(local_updates[0]["updates"]))
        ]
        for idx_weight in range(len(local_updates[0]["updates"])):
            aggregated_delta_weights[idx_weight] = sum(
                [
                    local_updates[idx_client]["updates"][idx_weight] * local_updates[idx_client]["n_samples"]
                    for idx_client in range(self.num_clients)
                ]
            )
            aggregated_delta_weights[idx_weight] /= float(self.total_number_of_samples)
        
        # 初始化全局模型
        global_model = self.models_list[0]
        local_previous_state_0 = global_model._get_current_params()
        global_model._update_params(aggregated_delta_weights)
        
        # fedavg下的全局模型精度
        pooled_perf_dict = evaluate_model_on_local_and_pooled_tests(global_model.model, self.test_dls, self.test_pooled, self.metric, self.evaluate_func,local =False)
        performance["avg_gloabl_model"] = pooled_perf_dict['client_test_0']
        print("avg_gloabl_model", performance["avg_gloabl_model"])
        
        # Reset local model
        for p_new, p_old in zip(global_model.model.parameters(), local_previous_state_0):
            p_new.data = torch.from_numpy(p_old).to(p_new.device)
        del local_previous_state_0
        
        # 测量每个节点的贡献
        for idx_client in range(self.num_clients):
            aggregated_delta_weights = [
                None for _ in range(len(local_updates[0]["updates"]))
            ]
            for idx_weight in range(len(local_updates[0]["updates"])):
                aggregated_delta_weights[idx_weight] = sum(
                    [
                        local_updates[idx]["updates"][idx_weight] * local_updates[idx]["n_samples"]
                        for idx in range(self.num_clients)
                    ]
                )
                aggregated_delta_weights[idx_weight] -= local_updates[idx_client]["updates"][idx_weight] * local_updates[idx_client]["n_samples"]
                aggregated_delta_weights[idx_weight] /= float(self.total_number_of_samples - local_updates[idx_client]["n_samples"])
            
            # 初始化全局模型
            global_model_test = self.models_list[0]
            local_previous_state_1 = global_model_test._get_current_params()
            global_model_test._update_params(aggregated_delta_weights)
  
            pooled_perf_dict = evaluate_model_on_local_and_pooled_tests(global_model_test.model, self.test_dls, self.test_pooled, self.metric, self.evaluate_func,local =False)
            performance[f"except_client_{idx_client}"] = pooled_perf_dict['client_test_0']
            print(f"except_client_{idx_client}: ",performance[f"except_client_{idx_client}"])
            
            # Reset local model
            for p_new, p_old in zip(global_model_test.model.parameters(), local_previous_state_1):
                p_new.data = torch.from_numpy(p_old).to(p_new.device)
            del local_previous_state_1
        
        best_except = []
        contribute_list = []
        contribute = 0.0
        for i in range(self.num_clients):
            contribute = performance[f"avg_gloabl_model"]-performance[f"except_client_{i}"]
            if contribute < -0.05:
                best_except.append(i)
                contribute = 0
            elif contribute < 0:
                contribute = contribute * (-0.07)
            contribute_list.append(contribute)
        contribute_sum = sum(contribute_list)
        contribute_list.append(contribute_sum)
        print("contribute:",contribute_list)
        self.writer1.writerow(contribute_list)
        
        # 根据贡献聚合全局模型
        aggregated_delta_weights = [
            0.0 for _ in range(len(local_updates[0]["updates"]))
        ]
        no_join = 0
        print("best_except",best_except)
        for id in best_except:
            print("id:",id)
            no_join += local_updates[id]["n_samples"]
        join_counts=self.total_number_of_samples- no_join
        for idx_weight in range(len(local_updates[0]["updates"])):
            for idx_client in range(self.num_clients):
                if idx_client not in best_except:
                    aggregated_delta_weights[idx_weight] += (local_updates[idx_client]["updates"][idx_weight] * float(self.alpha * local_updates[idx_client]["n_samples"]/join_counts + self.xishu * contribute_list[idx_client]/contribute_sum))
        
        # Update models
        for _model in self.models_list:
            _model._update_params(aggregated_delta_weights)
        
        print("i:",self.counts)
        torch.save(self.models_list[0].model.state_dict(), '/root/autodl-tmp/isic2019/FL_evolution/FL_evolution_60/fedcontribution/model_weights_'+str(self.counts)+'.pth')
        
        # 计算贡献全局模型的精度
        (
            perf_dict,
            pooled_perf_dict,
            y_true_dict,
            y_pred_dict,
            y_true_pooled_dict,
            y_pred_pooled_dict,
        ) = evaluate_model_on_local_and_pooled_tests(self.models_list[0].model, self.test_dls, self.test_pooled, self.metric, self.evaluate_func,local =True)
        #print("Global Model Per-center performance:")
        print(perf_dict)
        print("Global Model Performance on pooled test set: ",pooled_perf_dict['client_test_0'])
        
        # 写csv表
        global_model_performance = list(perf_dict.values())
        global_model_performance.append(pooled_perf_dict['client_test_0'])
        
        round_performance = list(performance.values()) + global_model_performance
        
        self.writer0.writerow(round_performance)
        
        
    def run(self,xishu):
        """This method performs self.nrounds rounds of averaging
        and returns the list of models.
        """
        self.alpha = round(1-xishu,2)
        self.s = str(self.alpha)+'_'+str(xishu)
        record_name = "/root/autodl-tmp/isic2019/FL_evolution_60/fedcontribution.csv"
        print("record_name:",record_name)
        contribution_name = "/root/autodl-tmp/isic2019/FL_evolution_60/contribution.csv"

        csvfile0 = open(record_name, 'w',newline='', encoding="utf-8")
        csvfile1 = open(contribution_name, 'w',newline='', encoding="utf-8")
        self.writer0 = csv.writer(csvfile0)
        self.writer1 = csv.writer(csvfile1)
        """
        self.writer0.writerow([
            'avg_pooled',
            'except_client0',
            'except_client1',
            'client0_fedcon',
            'client1_fedcon',
            'pooled_fedcon',
        ]) # 标题，写入一行
        self.writer1.writerow([
            'client0',
            'client1',
            'sum',
        ])
        
        self.writer0.writerow([
            'avg_pooled',
            'except_client0',
            'except_client1',
            'except_client2',
            'except_client3',
            'client0_fedcon',
            'client1_fedcon',
            'client2_fedcon',
            'client3_fedcon',
            'pooled_fedcon',
        ]) # 标题，写入一行
        self.writer1.writerow([
            'client0',
            'client1',
            'client2',
            'client3',
            'sum',
        ])
        
        self.writer0.writerow([
            'avg_pooled',
            'except_client0',
            'except_client1',
            'except_client2',
            'except_client3',
            'except_client4',
            'except_client5',
            'client0_fedcon',
            'client1_fedcon',
            'client2_fedcon',
            'client3_fedcon',
            'client4_fedcon',
            'client5_fedcon',
            'pooled_fedcon',
        ]) # 标题，写入一行
        self.writer1.writerow([
            'client0',
            'client1',
            'client2',
            'client3',
            'client4',
            'client5',
            'sum',
        ])
        """
        
        self.writer0.writerow([
            'avg_pooled',
            'except_client1',
            'except_client2',
            'Client1',
            'Client2',
            'Pooled',
        ]) # 标题，写入一行
        self.writer1.writerow([
            'Client1',
            'Client2',
            'Sum',
        ])
        
        self.xishu = xishu
        
        for i in tqdm(range(self.nrounds)):
            self.counts = i
            self.perform_round()
        return [m.model for m in self.models_list]
        