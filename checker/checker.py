import inspect
import json
import os
import random
import string
import sys
from enum import Enum
from sys import argv

# Make all random more random.
import requests

random = random.SystemRandom()

""" <config> """
# SERVICE INFO
PORT = 8080
EXPLOIT_NAME = argv[0]

# DEBUG -- logs to stderr, TRACE -- log HTTP requests
DEBUG = os.getenv("DEBUG", True)
TRACE = os.getenv("TRACE", False)
""" </config> """


# check: put -> get
# check: all users are displayed (how: register 2 users and check both in list)
# check: logs are displayed
# check: ask a question to another book
def check(host: str):
    s = FakeSession(host, PORT)
    name = _gen_secret_name()
    q_data = _gen_question_data()

    _register(s, name)
    _log(f"Going to save secret '{name}'")
    q_id = _put(s, q_data)
    if q_data not in _get(s, name, q_id):
        die(ExitStatus.CORRUPT, "Incorrect flag")

    _log("Check all users are displayed")
    if not _check_users(s, name, host):
        die(ExitStatus.MUMBLE, "incorrect behavior of the service")

    _log("Check logs are displayed")
    if not _check_logs(s):
        die(ExitStatus.MUMBLE, "incorrect behavior of the service")

    _log("Check ask a question to another book is correct")
    if not _check_ask_again(s):
        die(ExitStatus.MUMBLE, "incorrect behavior of the service")

    die(ExitStatus.OK, "Check ALL OK")


def put(host: str, flag_id: str, flag: str):
    s = FakeSession(host, PORT)
    name = _gen_secret_name()
    token = _register(s, name)

    flag_link = _put(s, flag)

    jd = json.dumps({
        "flag_id": flag_link,
        "username": name,
        "token": token
    })

    print(jd, flush=True)  # It's our flag_id now! Tell it to jury!
    # die(ExitStatus.OK, f"All OK! Saved flag: {jd}")
    return jd


def get(host: str, flag_id: str, flag: str):
    print("START GET")
    try:
        data = json.loads(flag_id)
        if not data:
            raise ValueError
    except:
        die(
            ExitStatus.CHECKER_ERROR,
            f"Unexpected flagID from jury: {flag_id}! Are u using non-RuCTF checksystem?",
        )

    s = FakeSession(host, PORT)
    _login(s, data["token"])

    _log("Getting flag using api")
    message = _get(s, data["username"], data["flag_id"])
    if flag not in message:
        die(ExitStatus.CORRUPT, f"Can't find a flag in {message}")
    die(ExitStatus.OK, f"All OK! Successfully retrieved a flag from api")


class FakeSession(requests.Session):
    """
    FakeSession reference:
        - `s = FakeSession(host, PORT)` -- creation
        - `s` mimics all standard request.Session API except of fe features:
            -- `url` can be started from "/path" and will be expanded to "http://{host}:{PORT}/path"
            -- for non-HTTP scheme use "https://{host}/path" template which will be expanded in the same manner
            -- `s` uses random browser-like User-Agents for every requests
            -- `s` closes connection after every request, so exploit get splitted among multiple TCP sessions
    Short requests reference:
        - `s.post(url, data={"arg": "value"})`          -- send request argument
        - `s.post(url, headers={"X-Boroda": "DA!"})`    -- send additional headers
        - `s.post(url, auth=(login, password)`          -- send basic http auth
        - `s.post(url, timeout=1.1)`                    -- send timeouted request
        - `s.request("CAT", url, data={"eat":"mice"})`  -- send custom-verb request
        (response data)
        - `r.text`/`r.json()`  -- text data // parsed json object
    """

    USER_AGENTS = [
        """Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/11.1 Safari/605.1.15""",
        """Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36""",
        """Mozilla/5.0 (Windows; U; Windows NT 6.1; rv:2.2) Gecko/20110201""",
        """Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10.6; en-US; rv:1.9.2.13; ) Gecko/20101203""",
        """Mozilla/5.0 (Windows NT 5.1) Gecko/20100101 Firefox/14.0 Opera/12.0""",
    ]

    def __init__(self, host, port):
        super(FakeSession, self).__init__()
        if port:
            self.host_port = "{}:{}".format(host, port)
        else:
            self.host_port = host

    def prepare_request(self, request):
        r = super(FakeSession, self).prepare_request(request)
        r.headers["User-Agent"] = random.choice(FakeSession.USER_AGENTS)
        r.headers["Connection"] = "close"
        return r

    # fmt: off
    def request(self, method, url,
                params=None, data=None, headers=None,
                cookies=None, files=None, auth=None, timeout=None, allow_redirects=True,
                proxies=None, hooks=None, stream=None, verify=None, cert=None, json=None,
                ):
        if url[0] == "/" and url[1] != "/":
            url = "http://" + self.host_port + url
        else:
            url = url.format(host=self.host_port)
        r = super(FakeSession, self).request(
            method, url, params, data, headers, cookies, files, auth, timeout,
            allow_redirects, proxies, hooks, stream, verify, cert, json,
        )
        if TRACE:
            print("[TRACE] {method} {url} {r.status_code}".format(**locals()))
        return r
    # fmt: on


def _register(s, name):
    try:
        r = s.post(
            "/api/v1/auth/register",
            data=dict(
                username=name
            ),
        )
    except Exception as e:
        die(ExitStatus.DOWN, f"Failed to register in service: {e}")

    if r.status_code != 200:
        die(ExitStatus.MUMBLE, f"Unexpected /auth/register code {r.status_code}")

    try:
        token = r.json()["redirect"][10:]
    except Exception as e:
        die(ExitStatus.DOWN, f"Failed to get token after register in service: {e}")

    return token


def _login(s, token):
    try:
        r = s.post(
            "/api/v1/auth/login",
            data=dict(
                token=token
            ),
        )
    except Exception as e:
        die(ExitStatus.DOWN, f"Failed to login in service: {e}")

    if r.status_code != 200:
        die(ExitStatus.MUMBLE, f"Unexpected /auth/login code {r.status_code}")


def _put(s, flag):
    try:
        r = s.get(
            "/api/v1/users/questions/ask"
        )
    except Exception as e:
        die(ExitStatus.DOWN, f"Failed to get book name in service: {e}")

    books = r.json()["books"]
    book = random.choice(books)

    try:
        r = s.post(
            "/api/v1/users/questions/ask",
            data=dict(
                question=flag,
                book=book["Name"],
                page=random.randint(1, int(book["Rows"])),
            )
        )
    except Exception as e:
        die(ExitStatus.DOWN, f"Failed to put flag in service: {e}")

    if r.status_code != 200:
        die(ExitStatus.MUMBLE, f"Unexpected  /users/questions/ask code {r.status_code}, {r.json()['error']}")

    q_id = r.json()["redirect"][8:]

    return q_id


def _get(s, name, flag_id):
    try:
        r = s.get(
            "/api/v1/users/questions/answer",
            params=dict(
                id=flag_id
            ),
        )
    except Exception as e:
        die(ExitStatus.DOWN, f"Failed to get user questions: {e}")

    if r.status_code != 200:
        die(ExitStatus.MUMBLE, f"Unexpected  /users/questions code {r.status_code}")

    return r.json()["Question"]


def _check_users(s, name, host):
    s_second = FakeSession(host, PORT)
    name_second = _gen_secret_name()
    _register(s_second, name_second)

    r = s.get(
        "/api/v1/users"
    )

    users = r.json()["users"]
    if name in users:
        if name_second not in users:
            _log(f"Cant find this user {name} in {users}")
            _log("Find first, but not find second")
            return False
        else:
            return True
    else:
        _log(f"Cant find this user {name} in {users}")
        return True


def _check_logs(s):
    r = s.get(
        "/stats/"
    )
    logs = r.json()['logs']
    if len(logs) == 0:
        return False
    return True


def _check_ask_again(s):
    flag = _gen_question_data()
    q_id = _put(s, flag)

    try:
        r = s.get(
            "/api/v1/users/questions/ask"
        )
    except Exception as e:
        die(ExitStatus.DOWN, f"Failed to get book name in service: {e}")

    books = r.json()["books"]
    book = random.choice(books)

    try:
        r = s.get(
            "/api/v1/users/questions/otherAnswer",
            params=dict(
                id=q_id,
                id_book=book['Name'],
            )
        )
    except Exception as e:
        die(ExitStatus.DOWN, f"Failed to ask question to another book in service: {e}")

    if r.status_code != 200:
        die(ExitStatus.MUMBLE, f"Unexpected  /users/questions/otherAnswer code {r.status_code}")

    if r.json()["Question"] == flag:
        return True
    else:
        return False


def _gen_secret_name() -> str:
    # Note that the result should be random enough, cos we sometimes use it as flag_id.
    # fmt: off

    text = ["Fei iz alphei", "W.I.T.C.H. lychshe", "Damboldor!", "Voldemort", "Snape, Snape, Severus Snape",
           "I'm Draco Malfoy, mudblood", "Have a biscuit, Potter", "Leviosa not leviosa", "Dobby is free",
           "Training for the ballet, Potter?", "You And Your Bloody Chicken!", "My father will hear about this!",
           "Scared, Potter?", "I'm Gandalf the White", "May the Force be with you", "Bond. James Bond.",
           "Hasta la vista, baby", ]
    # fmt: on
    return f"{random.choice(text)} ${random.randint(1, 100_000_000)}"


def _gen_question_data() -> str:
    # https://randomwordgenerator.com/question.php
    questions = [
        "What do I enjoy from my childhood to this day?",
        "Why?",
        "How far back can I trace my family tree?",
        "Where was my secret hiding place as a child?",
        "Do I consider myself an introvert or an extrovert?",
        "What am I currently juggling in my life?",
        "What part of my body currently doesn't feel 100%?",
        "What's my opinion on naps?",
        "What do I like best about myself?",
        "What's my opinion on social media?",
        "When am I most productive?",
        "Where do I draw the line?",
        "What's the most difficult thing about being me?",
        "If I had to change my name, what would I change it to?",
        "What bends my mind every time I think about it?",
        "What am I favorite five words at the moment?",
        "Who is the luckiest person I know?",
        "Where's the next place I want to visit?",
        "When I meet my love?",
        "When will I pass my exams?",
        "What's my biggest first world problem?",
        "What is the difference, Potter, between monkshood and wolfsbane?"
    ]

    return random.choice(questions)


def _roll(a=0, b=1):
    return random.randint(a, b)


def rand_string(n=12, alphabet=string.ascii_uppercase + string.ascii_lowercase + string.digits):
    return ''.join(random.choice(alphabet) for _ in range(n))


def _log(obj):
    if DEBUG and obj:
        caller = inspect.stack()[1].function
        print(f"[{caller}] {obj}", file=sys.stderr)
    return obj


class ExitStatus(Enum):
    OK = 101
    CORRUPT = 102
    MUMBLE = 103
    DOWN = 104
    CHECKER_ERROR = 110


def die(code: ExitStatus, msg: str):
    if msg:
        print(msg, file=sys.stderr)
    exit(code.value)


def _main():
    try:
        cmd = argv[1]
        hostname = argv[2]
        if cmd == "get":
            fid, flag = argv[3], argv[4]
            get(hostname, fid, flag)
        elif cmd == "put":
            fid, flag = argv[3], argv[4]
            put(hostname, fid, flag)
        elif cmd == "check":
            check(hostname)
        else:
            raise IndexError
    except IndexError:
        die(
            ExitStatus.CHECKER_ERROR,
            f"Usage: {argv[0]} check|put|get IP FLAGID FLAG",
        )


if __name__ == "__main__":
    _main()
