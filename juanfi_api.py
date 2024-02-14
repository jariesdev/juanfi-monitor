import time
import requests
from requests import Response
from collections import deque
from datetime import datetime


class JuanfiApi:
    _system_uptime_ms: int = 0
    _baseUrl = "http://192.168.42.10:8081"
    _row_separator = "|"

    def run(self):
        self._load_system_uptime()
        print(self.get_system_logs(), str(self._system_uptime_ms))

    def get_system_logs(self) -> list:
        self._load_system_uptime()
        r = self._send_api_request("api/getSystemLogs")
        if r.status_code != 200:
            raise Exception("Unable to retrieve system logs")

        if not r.text:
            return []

        rows = []
        _r = r.text.split(self._row_separator)
        for _c in _r:
            header_data = _c.split("~")
            if len(header_data) > 1:
                _c_data = header_data[1]
            else:
                _c_data = header_data[0]

            parts = deque(_c_data.split("#"))
            log_time = parts.popleft()
            col = [
                self._compute_log_time(int(log_time)),
                self._format_log_message(list(parts))
            ]
            rows.append(col)
        rows.reverse()
        return rows

    def _load_system_uptime(self) -> None:
        r = self._send_api_request("api/dashboard")
        if r.status_code != 200:
            raise Exception("Unable to retrieve system uptime")
        body = r.text
        uptime = body.split(self._row_separator)[0]
        self._system_uptime_ms = int(uptime)

    def _send_api_request(self, url: str) -> Response:
        timestamp = self._get_current_milli_time()
        url = ("%s/admin/%s?query=%d" % (self._baseUrl, url, timestamp))
        return requests.get(url=url, headers={'X-TOKEN': self.get_api_key()}, timeout=5)

    def _get_current_milli_time(self) -> int:
        return round(time.time() * 1000)

    def get_api_key(self) -> str:
        return "55eav610vk"

    def _compute_log_time(self, timeSinceStartup: int) -> str:
        uptime = self._system_uptime_ms
        dt = datetime.fromtimestamp(((time.time() * 1000 - int(uptime)) + timeSinceStartup) / 1000)
        return dt.strftime("%Y-%m-%d %H:%M:%S")

    def _format_log_message(self, data: list) -> str:
        types = self._get_log_types()
        parts = deque(data)
        type_index = int(parts.popleft())
        template = types[type_index]
        return template.format(*list(parts))

    def _get_log_types(self) -> list:
        log_types = []
        log_types.append('No log message')
        log_types.append('System Starting...')
        log_types.append('Network Connected Succesfully')
        log_types.append('Mikrotik Connected Succesfully')
        log_types.append('Juanfi Initial Setup')
        log_types.append('Failed to login to mikrotik')
        log_types.append('{0} Cancel Topup')
        log_types.append('Rates modified')
        log_types.append('System Configuration Modified')
        log_types.append('Reset Total Sales')
        log_types.append('Reset Current sales')
        log_types.append('Reset Customer Count')
        log_types.append('{0} Login Sucessfully')
        log_types.append('{0} Login Failed')
        log_types.append('{0} Purchase {1}, amount: {2}')
        log_types.append('{0} Attempted to insert coin')
        log_types.append('{0} was banned from using coinslot')
        log_types.append('Create voucher failed {0}, retrying...')
        log_types.append('{0} Inserted coin {1}')
        log_types.append('Manual voucher purchase activated')
        log_types.append('Generated {0} voucher(s)')
        log_types.append('NightLight Turn On')
        log_types.append('NightLight Turn Off')
        log_types.append('Kick active user {0}')
        log_types.append('{0} tried to insert coin but no internet available')
        log_types.append('Reset Daily Sales')
        log_types.append('Reset Monthly Sales')
        log_types.append('Update charging settings')
        log_types.append('{0} Purchase Eload success {1}')
        log_types.append('Update Eload settings')
        log_types.append('Update Eload rates')
        log_types.append('Clear Eload transactions')
        log_types.append('{0} Purchase Eload failed {1}')
        log_types.append('Unusual coinslot pulse detected, please check coinslot')
        return log_types


if __name__ == '__main__':
    logger = JuanfiApi()
    logger.run()
