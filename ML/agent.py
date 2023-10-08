import math
import statistics
import torch
import model_2

device = torch.device("cuda" if torch.cuda.is_available() else "cpu")


class Agent(object):
    def __init__(self, args, eva_dataset, global_model):
        if args['dataset'] == 'mnist':
            self.model = model_2.MNIST_CNN_Net()
        else:
            self.model = model_2.CIFAR_CNN_Net()
        self.eva_loader = torch.utils.data.DataLoader(eva_dataset, batch_size=args['batch_size'], shuffle=True)
        # self.model.load_state_dict(torch.load(model_path))
        self.model.load_state_dict(global_model)
        self.model.to(device)

    def evaluate(self):
        self.model.to(device)
        self.model.eval()
        correct = 0
        data_size = 0
        loss = 0
        Loss = torch.nn.CrossEntropyLoss()
        for batch_id, batch in enumerate(self.eva_loader):
            data, target = batch
            data_size += data.size()[0]
            data = data.to(device)
            target = target.to(device)

            output = self.model(data)
            loss += Loss(output, target)

            pred = output.data.max(1)[1]
            correct += pred.eq(target.data.view_as(pred)).cpu().sum().item()

        eval_acc = 100.0 * (float(correct) / float(data_size))
        avg_loss = loss / data_size

        return eval_acc, avg_loss.cpu().item()


def cal_diff(validate_loss_list, train_loss):
    beta = 1
    gama = 1
    diff_i = 0

    loss_im = median = statistics.median(validate_loss_list)
    for loss_ij in validate_loss_list:
        diff_i += gama * abs(loss_ij - train_loss) / math.exp(beta * (loss_ij - loss_im))

    return diff_i


def cal_agg_weight(diff_list, loss_list, data_num_list):
    trainer_num = len(diff_list)

    sort_loss_list = sorted(loss_list)
    median_loss = sort_loss_list[int(len(loss_list) / 2)]
    abs_loss = [abs(i - median_loss) for i in loss_list]
    sort_abs_loss = sorted(abs_loss)
    median_abs_loss = sort_abs_loss[int(len(abs_loss) / 2)]
    loss_MAD = 3.5 * median_abs_loss

    sort_diff_list = sorted(diff_list)
    median_diff = sort_diff_list[int(len(diff_list) / 2)]
    abs_diff = [abs(j - median_diff) for j in diff_list]
    sort_abs_diff = sorted(abs_diff)
    median_abs_diff = sort_abs_diff[int(len(abs_diff) / 2)]
    diff_MAD = 3.9 * median_abs_diff

    Di_list = []
    test_di = []
    for i in range(trainer_num):
        dist = abs(diff_list[i] - median_diff)
        test_di.append(dist)
        if dist <= diff_MAD:
            Di_list.append(diff_MAD - dist)
        else:
            Di_list.append(0)

    Li_list = []
    test_li = []
    for i in range(trainer_num):
        dist = abs(loss_list[i] - median_loss)
        test_li.append(dist)
        if dist <= loss_MAD:
            Li_list.append(loss_MAD - dist)
        else:
            Li_list.append(loss_MAD - dist)

    Si_list = []
    for i in range(trainer_num):
        Si_list.append(Di_list[i] * Li_list[i])

    weighted_data_num_sum = 0
    for i in range(trainer_num):
        weighted_data_num_sum += data_num_list[i] * Si_list[i]

    agg_weights = []
    for i in range(trainer_num):
        agg_weights.append(data_num_list[i] * Si_list[i] / weighted_data_num_sum)

    return agg_weights, test_di, test_li, Si_list
