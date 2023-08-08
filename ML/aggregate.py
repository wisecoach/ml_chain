import tornado.web


class AggregateHandler(tornado.web.RequestHandler):

    def get(self, task_id):
        print(task_id)

    def post(self, task_id):
        print(task_id)
        weight = self
