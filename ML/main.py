import argparse
import task as t
from typing import Any

import tornado.ioloop
import tornado
import json

from tornado import httputil


class TaskHandler(tornado.web.RequestHandler):

    task_dict = {}

    # 10Task
    # 10个client
    def post(self, path):
        print(path)
        task_id = path.split("/")[-1]
        if path.__contains__("init"):
            self.init(task_id)
        if path.__contains__("train"):
            self.train(task_id)
        if path.__contains__("validate"):
            self.validate(task_id)
        if path.__contains__("aggregate"):
            self.aggregate(task_id)

    def init(self, task_id):
        genesis = json.loads(self.request.body)
        new_task = t.Task(genesis)
        TaskHandler.task_dict[task_id] = new_task
        print("init success")

    def train(self, task_id):
        req = json.loads(self.request.body)
        resp = TaskHandler.task_dict[task_id].train(req)
        print("train success")
        self.write(json.dumps(resp))

    def validate(self, task_id):
        req = json.loads(self.request.body)
        resp = TaskHandler.task_dict[task_id].validate(req)
        print("validate success")
        self.write(json.dumps(resp))

    def aggregate(self, task_id):
        req = json.loads(self.request.body)
        resp = TaskHandler.task_dict[task_id].aggregate(req)
        print("aggregate success")
        self.write(json.dumps(resp))


def make_app():
    return tornado.web.Application([
        (r"/(.*)", TaskHandler,),
    ])


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('--port', type=int, default=10099, help="port we use")
    args = parser.parse_args()
    app = make_app()
    app.listen(args.port)
    tornado.ioloop.IOLoop.current().start()
