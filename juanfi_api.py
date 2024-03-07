import time
from typing import Union
from datetime import datetime
import http.client
from http.client import HTTPResponse

from sql_app.models import Vendo
from urllib.parse import urlparse

STATS_TYPE_COIN_COUNT = "coinCount"


class JuanfiApi():
    _system_uptime_ms: int = 0
    _base_url = "http://192.168.42.10:8081"
    _api_key = ""
    _row_separator = "|"
    system_logs = []

    def __init__(self, vendo: Vendo):
        self._base_url = vendo.api_url
        self._api_key = vendo.api_key

    def run(self):
        self.load_system_logs()
        print(self.get_formatted_system_logs())

    def load_system_logs(self) -> list:
        self._load_system_status()
        r = self._send_api_request("api/getSystemLogs")
        _body = r.read().decode()
        if r.status != 200:
            raise Exception("Unable to retrieve system logs. Error code: {}".format(r.status))
        if not _body or _body == "":
            return []

        rows = []
        _r = _body.split(self._row_separator)
        for _c in _r:
            has_header = False
            header_data = _c.split("~")
            if len(header_data) > 1:
                _c_data = header_data[1]
                has_header = True
            else:
                _c_data = header_data[0]
            _i = _c_data.split("#")

            log_params = []
            if len(_i) >= 3:
                # take parameters from index 2 to the end
                log_params = _i[2:]

            rows.append({
                "has_header": has_header,
                "time": int(_i[0]),
                "log_type_index": int(_i[1]),
                "log_params": log_params,
            })
        rows.reverse()
        self.system_logs = rows
        return rows

    def get_formatted_system_logs(self) -> list:
        rows = []
        _r = self.system_logs
        for _c in _r:
            col = [
                self.compute_log_time(_c.get("time")),
                self._format_log_message(_c.get("log_type_index"), _c.get("log_params"))
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
            "server_time": time.time() * 1000,
        }

    def _load_system_status(self) -> None:
        r = self._send_api_request("api/dashboard")
        _body = r.read().decode()
        if r.status != 200:
            raise Exception("Failed to retrieve system status. Error code: {}".format(r.status))

        data = _body.split(self._row_separator)
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

    def _send_api_request(self, url: str, method: str = "GET", query: Union[dict, None] = None) -> HTTPResponse:
        try:
            p = urlparse(self._base_url)
            conn = http.client.HTTPConnection(host=p.hostname, port=p.port, timeout=5)
            timestamp = self._get_current_milli_time()

            query_params = query if query is not None else {}
            query_params['query'] = timestamp
            url_params = []
            for key, value in query_params.items():
                url_params.append(f"{key}={value}")

            url = ("/admin/%s?%s" % (url, '&'.join(url_params)))
            conn.request(method=method, url=url, headers={'X-TOKEN': self.get_api_key()})
            response = conn.getresponse()
            conn.close()
            return response
        except Exception as e:
            raise Exception("Error while connecting to API")

    def _get_current_milli_time(self) -> int:
        return round(time.time() * 1000)

    def get_api_key(self) -> str:
        return self._api_key

    def compute_log_time(self, time_since_startup: int, precision: Union[str, None] = None) -> str:
        __format = "%Y-%m-%d %H:%M:%S"
        if precision is not None and precision == "M":
            __format = "%Y-%m-%d %H:%M:00"

        uptime = self._system_uptime_ms
        dt = datetime.fromtimestamp(((time.time() * 1000 - int(uptime)) + time_since_startup) / 1000)

        return dt.strftime(__format)

    def _format_log_message(self, log_type: int, params: list) -> str:
        types = self._get_log_types()
        template = types[log_type]
        return template.format(*params)

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

    def get_vendo_status(self) -> dict:
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
            "server_time": time.time() * 1000,
        }

    def reset_current_sales(self) -> None:
        stat_type = STATS_TYPE_COIN_COUNT
        query = {'type': stat_type}
        r = self._send_api_request(url="api/resetStatistic", query=query)
        _body = r.read().decode()
        if r.status != 200:
            raise Exception("Unable to reset current sales. Error code: {}".format(r.status))


if __name__ == '__main__':
    logger = JuanfiApi()
    logger.run()
