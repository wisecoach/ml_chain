import torch
import numpy as np
from torch import nn
import torch.nn.functional as F


class MNIST_CNN_Net(nn.Module):
    def __init__(self):
        super(MNIST_CNN_Net, self).__init__()
        self.conv1 = nn.Conv2d(1, 20, 5, 1)
        self.conv2 = nn.Conv2d(20, 50, 5, 1)
        self.fc1 = nn.Linear(5 * 5 * 50, 500)
        self.fc2 = nn.Linear(500, 10)

    def forward(self, x):
        x = F.relu(self.conv1(x))
        x = F.max_pool2d(x, 2, 2)
        x = F.relu(self.conv2(x))
        x = F.max_pool2d(x, 2, 2)
        x = x.view(-1, 5 * 5 * 50)
        x = F.relu(self.fc1(x))
        x = self.fc2(x)
        return F.log_softmax(x, dim=1)

    def reshape(self, flat_gradient):
        conv1_weight = 20 * 1 * 5 * 5
        conv1_bias = 20
        conv2_weight = 50 * 20 * 5 * 5
        conv2_bias = 50
        fc1_weight = 500 * 1250
        fc1_bias = 500
        fc2_weight = 10 * 500
        fc2_bias = 10
        param_num_list = [conv1_weight, conv1_bias, conv2_weight, conv2_bias, fc1_weight, fc1_bias, fc2_weight, fc2_bias]
        param_tuple_list = [
            (20, 1, 5, 5,),
            (20,),
            (50, 20, 5, 5,),
            (50,),
            (500, 1250,),
            (500,),
            (10, 500,),
            (10,),
        ]
        layers = []
        right = 0
        for i in range(8):
            left = right
            right = right + param_num_list[i]
            layers.append(torch.from_numpy(np.reshape(flat_gradient[left:right], param_tuple_list[i])).type(torch.FloatTensor))
        return layers


class CIFAR_CNN_Net(torch.nn.Module):
    def __init__(self):
        super(CIFAR_CNN_Net, self).__init__()
        self.conv1 = nn.Conv2d(3, 32, 5, 1)
        self.conv2 = nn.Conv2d(32, 64, 5, 1)
        self.fc1 = nn.Linear(1600, 512)
        self.fc2 = nn.Linear(512, 10)

    def forward(self, x):
        x = F.relu(self.conv1(x))
        x = F.max_pool2d(x, 2, 2)
        x = F.relu(self.conv2(x))
        x = F.max_pool2d(x, 2, 2)
        x = x.view(-1, 1600)
        x = F.relu(self.fc1(x))
        x = self.fc2(x)
        # return x
        return F.log_softmax(x, dim=1)

    def reshape(self, flat_gradient):
        layers = [torch.from_numpy(np.reshape(flat_gradient[0:self.D_in * self.D_out], (self.D_out, self.D_in))).type(
            torch.FloatTensor),
                  torch.from_numpy(flat_gradient[self.D_in * self.D_out:self.D_in * self.D_out + self.D_out]).type(
                      torch.FloatTensor)]
        return layers
