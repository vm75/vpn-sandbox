from .openvpn import init_module as init_openvpn, shutdown as openvpn_shutdown
from .wireguard import init_module as init_wireguard, shutdown as wireguard_shutdown
from .proxy import init_proxy_module, HttpProxy, SocksProxy
