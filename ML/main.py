import tornado.ioloop
import tornado.web
import json


class MainHandler(tornado.web.RequestHandler):
    # 10Task
    # 10个client
    def get(self):
        # wt_1, losst_1 = clients[task].train_one_epoch(Wt)
        self.write("wt_1, losst_1")


class UserHandler(tornado.web.RequestHandler):
    users = [
        {"id": 1, "name": "Alice"},
        {"id": 2, "name": "Bob"},
        {"id": 3, "name": "Charlie"}
    ]

    def get(self, user_id=None):
        if user_id:
            for user in self.users:
                if user["id"] == int(user_id):
                    self.write(json.dumps(user))
                    return
            self.set_status(404)
        else:
            self.write(json.dumps(self.users))

    def post(self):
        user = json.loads(self.request.body)
        user["id"] = len(self.users) + 1
        self.users.append(user)
        self.write(json.dumps(user))

    def put(self, user_id=None):
        if user_id:
            for user in self.users:
                if user["id"] == int(user_id):
                    updated_user = json.loads(self.request.body)
                    updated_user["id"] = user_id
                    self.users[self.users.index(user)] = updated_user
                    self.write(json.dumps(updated_user))
                    return
            self.set_status(404)
        else:
            self.set_status(400)

    def delete(self, user_id=None):
        if user_id:
            for user in self.users:
                if user["id"] == int(user_id):
                    self.users.remove(user)
                    self.write("User with id %s deleted" % user_id)
                    return
            self.set_status(404)
        else:
            self.set_status(400)


def make_app():
    return tornado.web.Application([
        (r"/", MainHandler),
        (r"/users", UserHandler),
        (r"/users/(\d+)", UserHandler),
    ])


if __name__ == "__main__":
    app = make_app()
    app.listen(8888)
    tornado.ioloop.IOLoop.current().start()
