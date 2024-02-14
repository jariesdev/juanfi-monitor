import time
import requests
from requests import Response
from collections import deque
from datetime import datetime
import http.client
from http.client import HTTPResponse

class JuanfiApi:
    _system_uptime_ms: int = 0
    _baseUrl = "http://192.168.42.10:8081"
    _row_separator = "|"

    def run(self):
        print(self.get_system_logs(), self.get_system_status())

    def get_system_logs(self) -> list:
        self._load_system_status()
        r = self._send_api_request("api/getSystemLogs")
        if r.status != 200:
            raise Exception("Unable to retrieve system logs. Error code: {}".format(r.status))
        if not r.read().decode():
            return []

        rows = []
        _r = r.read().decode().split(self._row_separator)
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

    def get_system_status(self) -> dict:
        self._load_system_status()
        return {
            "system_uptime_ms": self._system_uptime_ms,
            "total_coin_count": self._total_coin_count,
            "current_coin_count": self._current_coin_count,
            "customer_count": self._customer_count,
            "internet_status": self._internet_status,
            "mikrotik_status": self._mikrotik_status,
            "mac_address": self._mac_address,
            "ip_address": self._ip_address,
            "hardware_type": self._hardware_type,
            "version": self._version,
            "interface_type": self._interface_type,
            "wireless_signal_strength": self._wireless_signal_strength,
            "free_heap": self._free_heap,
            "auth_type": self._auth_type,
            "night_light_status": self._night_light_status,
            "active_user_count": self._active_user_count,
            "system_clock": self._system_clock,
        }

    def _load_system_status(self) -> None:
        r = self._send_api_request("api/dashboard")
        if r.status != 200:
            raise Exception("Failed to retrieve system status. Error code: {}".format(r.status))

        data = r.read().decode().split(self._row_separator)
        self._system_uptime_ms = int(data[0])
        self._total_coin_count = int(data[1])
        self._current_coin_count = int(data[2])
        self._customer_count = int(data[3])
        self._internet_status = bool(data[4])
        self._mikrotik_status = bool(data[5])
        self._mac_address = str(data[6])
        self._ip_address = str(data[7])
        self._hardware_type = str(data[8])
        self._version = float(data[9])
        self._interface_type = str(data[10])
        self._wireless_signal_strength = int(data[11])
        self._free_heap = int(data[12])
        self._auth_type = str(data[13])
        self._night_light_status = bool(data[14])
        self._active_user_count = int(data[15])
        self._system_clock = str(data[16])

    def _send_api_request2(self, url: str) -> Response:
        timestamp = self._get_current_milli_time()
        url = ("%s/admin/%s?query=%d" % (self._baseUrl, url, timestamp))
        return requests.get(url=url, headers={'X-TOKEN': self.get_api_key()}, timeout=5)

    def _send_api_request(self, url: str) -> HTTPResponse:
        conn = http.client.HTTPConnection(host="192.168.42.10", port=8081, timeout=5)
        timestamp = self._get_current_milli_time()
        url = ("/admin/%s?query=%d" % (url, timestamp))
        conn.request(method="GET", url=url, headers={'X-TOKEN': self.get_api_key()})
        response = conn.getresponse()
        conn.close()
        return response

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
