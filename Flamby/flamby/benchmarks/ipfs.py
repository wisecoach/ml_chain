import io
import json

import ipfshttpclient as ipfsclient
import torch


def save_model(model_state_dict):
    """
        上传模型的权重到ipfs中
    """
    client = connect()
    buffer = io.BytesIO()
    torch.save(model_state_dict, buffer)
    return client.add_bytes(buffer.getvalue())


def load_model(model_hash):
    """
        读取ipfs中的模型权重
    """
    client = connect()
    model_bytes = client.cat(model_hash)
    buffer = io.BytesIO(model_bytes)
    return torch.load(buffer)


def save_update(updates):
    """
        存储update到ipfs中
    """
    client = connect()
    return client.add_json(updates)


def load_update(update_hash):
    """
        读取ipfs中的update
    """
    client = connect()
    update_bytes = client.cat(update_hash)
    buffer = io.BytesIO(update_bytes)
    return json.load(buffer)


def remove_model(model_hash):
    """
        删除ipfs中的模型权重
    """
    client = connect()
    model_bytes = client.remove(model_hash)
    buffer = io.BytesIO(model_bytes)
    return torch.load(buffer)


def connect():
    return ipfsclient.connect('/ip4/0.0.0.0/tcp/5001/http')
