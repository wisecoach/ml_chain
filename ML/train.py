import json

import tornado


class TrainHandler(tornado.web.RequestHandler):

    def get(self, task_id):
        print(task_id)

    def post(self, task_id=None):
        # global_weight = json.loads(self.request.body)
        print(task_id)
