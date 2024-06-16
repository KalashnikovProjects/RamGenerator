import queue
import threading
import time


class QueTimoutError(TimeoutError):
    pass


def _background_rate_limit_que_process(rate_que, last_request_time: list[int], rate_limit_time):
    while True:
        count_free = 0
        for j in last_request_time:
            if j + rate_limit_time <= time.time():
                count_free += 1
        sleep_count = min(rate_que.qsize(), count_free)
        for _ in range(sleep_count):
            event = rate_que.get()
            event.set()
            rate_que.task_done()

        time.sleep(0.1)


def api_rate_limiter_with_que(rate_limit: int, timeout: int = 200, rate_limit_time: int = 60):
    """
    Декоратор, устанавливает для функции рейт лимит с очередью

    Аргументы декоратора:
    timeout - таймаут нахождения элемента в очереди
    rate_limit - лимит вызовов функции,
    rate_limit_time=60 - лимит за какое время

    Задачи на которые не хватило лимита откладываются в очередь, элементы очереди -
    threading.Event(), евенты подаются в отдельном треде, wrapper функция ждёт евента и запускается.
    """
    rate_que = queue.Queue()
    last_request_time = [time.time() - rate_limit_time for _ in range(rate_limit)]

    threading.Thread(target=_background_rate_limit_que_process, daemon=True, args=(rate_que, last_request_time, rate_limit_time)).start()

    def que_rate_limit_decorator(func):
        def wrapper(*args, **kwargs):
            last_req_time_index, last_req_time = min(enumerate(last_request_time), key=lambda x: x[1])

            if not rate_que.empty() or last_req_time + rate_limit_time > time.time():
                event = threading.Event()
                rate_que.put(event)
                is_set = event.wait(timeout=timeout)
                if not is_set:
                    raise QueTimoutError("Таймаут очереди запросов")
            last_request_time[last_req_time_index] = time.time()
            res = func(*args, **kwargs)
            return res
        return wrapper

    return que_rate_limit_decorator
