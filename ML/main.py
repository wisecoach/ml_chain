import sys
import time

import task_normal as t
import tornado.ioloop
import tornado
import json
from concurrent.futures import ThreadPoolExecutor


class Executor(ThreadPoolExecutor):
    """ 单例模式
    """
    _instance = None

    def __new__(cls, *args, **kwargs):
        if not getattr(cls, '_instance', None):
            thread_num = 10 # 线程池数量
            cls._instance = ThreadPoolExecutor(max_workers=thread_num)
        return cls._instance


class TaskHandler(tornado.web.RequestHandler):

    task_dict = {}
    executor = Executor()

    def prepare(self):
        self.request.connection.set_max_body_size(9999999999999)

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
        print("ok")

    def init(self, task_id):
        genesis = json.loads(self.request.body)
        new_task = t.Task(genesis)
        TaskHandler.task_dict[task_id] = new_task
        print("init success")

    def train(self, task_id):
        print("train begin")
        print(time.strftime('%Y-%m-%d %H:%M:%S', time.localtime()))
        req = json.loads(self.request.body)
        resp = TaskHandler.task_dict[task_id].train(req)
        print("train success")
        print(time.strftime('%Y-%m-%d %H:%M:%S', time.localtime()))
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
    # parser = argparse.ArgumentParser()
    # parser.add_argument('--port', type=int, default=10099, help="port we use")
    # args = parser.parse_args()
    port = sys.argv[1]
    print("python server listen at: " + str(port))
    app = make_app()
    server = tornado.httpserver.HTTPServer(app, max_body_size=9999999999)
    server.listen(port)
    tornado.ioloop.IOLoop.current().start()
