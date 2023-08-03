import tornado.web


class ValidateHandler(tornado.web.RequestHandler):

    def get(self, task_id):
        print(task_id)

    def post(self, task_id):
        print(task_id)