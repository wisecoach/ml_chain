import copy
import random
import time

import torch


class Verifier(object):
    def __init__(self, args, eval_data):
        self.args = args
        self.verifier_id = id

        all_range = list(range(len(eval_data)))
        random.shuffle(all_range)
        data_len = 1000
        indices = all_range[0: data_len]

        self.eval_loader = torch.utils.data.DataLoader(eval_data, batch_size=self.args['batch_size'],
                                                       sampler=torch.utils.data.sampler.SubsetRandomSampler(indices))

    def test(self, test_model, device):
        test_model.to(device)
        test_model.eval()

        Loss = torch.nn.CrossEntropyLoss()
        total_loss = 0.0
        correct = 0
        data_size = 0
        for batch_id, batch in enumerate(self.eval_loader):
            time.sleep(0.012)
            data, target = batch
            data, target = data.to(device), target.to(device)
            data_size += data.size()[0]

            output = test_model(data)
            total_loss += Loss(output, target).item()
            pred = output.data.max(1)[1]
            correct += pred.eq(target.data.view_as(pred)).cpu().sum().item()

        acc = 100.0 * (float(correct) / float(data_size))
        test_loss = total_loss / data_size

        return test_loss, acc


