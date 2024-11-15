<template>
  <div id="modal"></div>
  <section class="section">
    <div class="container">
      <h1 class="title">VPN Sandbox</h1>

      <!-- Tabs -->
      <div class="tabs is-boxed">
        <ul>
          <li :class="{ 'is-active': currentTab === 'config' }">
            <a @click="currentTab = 'config'">Sandbox Config</a>
          </li>
          <li :class="{ 'is-active': currentTab === 'openvpn' }">
            <a @click="currentTab = 'openvpn'">OpenVPN Servers</a>
          </li>
          <li :class="{ 'is-active': currentTab === 'wireguard' }">
            <a @click="currentTab = 'wireguard'">Wireguard Servers</a>
          </li>
        </ul>
      </div>

      <!-- Sandbox Config -->
      <div v-if="currentTab === 'config'">
        <div class="columns">
          <div class="container column">
            <div class="container box" style="height: 100%;">
              <form>
                <div class="field is-horizontal">
                  <div class="field-label is-normal">
                    <legend class="label">VPN</legend>
                  </div>
                  <div class="field-body">
                    <div class="field control">
                      <basic id="vpn-switch" type="switch" v-model:value="openvpn.config.enabled"
                        @update:value="toggleModule('vpn')">
                      </basic>
                    </div>
                  </div>
                </div>
                <div class="field is-horizontal">
                  <div class="field-label is-normal">
                    <legend class="label">Http Proxy</legend>
                  </div>
                  <div class="field-body">
                    <div class="field control">
                      <basic id="http-proxy-switch" type="switch" v-model:value="http_proxy.config.enabled"
                        @update:value="toggleModule('http_proxy')">
                      </basic>
                    </div>
                  </div>
                </div>
                <div class="field is-horizontal">
                  <div class="field-label is-normal">
                    <legend class="label">Socks Proxy</legend>
                  </div>
                  <div class="field-body">
                    <div class="field control">
                      <basic id="socks-proxy-switch" type="switch" v-model:value="socks_proxy.config.enabled"
                        @update:value="toggleModule('socks_proxy')">
                      </basic>
                    </div>
                  </div>
                </div>
                <div>
                  <div class="divider">Common Config</div>
                </div>
                <div class="field is-horizontal">
                  <div class="field-label is-normal">
                    <legend class="label">LAN Subnets</legend>
                  </div>
                  <div class="field-body">
                    <div class="field control is-fullwidth">
                      <inline-list id="lan-subnets" :name="'Subnet'" v-model:entries="global.config.subnets"
                        type="subnet" @update:entries="setModified('global')">
                      </inline-list>
                    </div>
                  </div>
                </div>
                <div class="field is-horizontal">
                  <div class="field-label is-normal">
                    <legend class="label">VPN Type</legend>
                  </div>
                  <div class="field-body">
                    <div class="field">
                      <div class="control select is-fullwidth">
                        <select id="vpn-type" v-model="global.config.vpnType" @change="setModified('global')">
                          <option v-for="vpnType in global.config.vpnTypes" :key="vpnType" :value="vpnType"
                            :selected="vpnType === 'openvpn'">
                            {{ vpnType }}
                          </option>
                        </select>
                      </div>
                    </div>
                  </div>
                </div>
                <div v-if="global.config.vpnType === 'openvpn'">
                  <div>
                    <div class="divider">VPN Config</div>
                  </div>
                  <div class="field is-horizontal">
                    <div class="field-label is-normal">
                      <legend class="label">OpenVPN Provider</legend>
                    </div>
                    <div class="field-body">
                      <div class="field">
                        <div class="control select is-fullwidth">
                          <select id="openvpn-provider" v-model="openvpn.config.serverName"
                            @change="setModified('openvpn')">
                            <option v-for="server in openvpn.servers" :key="server.name" :value="server.name"
                              :selected="server.name === openvpn.config.serverName">
                              {{ server.name }}
                            </option>
                          </select>
                        </div>
                      </div>
                    </div>
                  </div>
                  <div class="field is-horizontal">
                    <div class="field-label is-normal">
                      <legend class="label">Server Endpoint</legend>
                    </div>
                    <div class="field-body">
                      <div class="field control select is-fullwidth">
                        <select id="openvpn-endpoint" v-model="openvpn.config.serverEndpoint"
                          @change="setModified('openvpn')">
                          <option v-for="endpoint in endpoints" :key="endpoint.name" :value="endpoint.name"
                            :selected="endpoint.name === openvpn.config.serverEndpoint">
                            {{ endpoint.name }}
                          </option>
                        </select>
                      </div>
                    </div>
                  </div>
                </div>
                <div v-if="global.config.vpnType === 'wireguard'">
                  <div>
                    <div class="divider">VPN Config</div>
                  </div>
                  <div class="field is-horizontal">
                    <div class="field-label is-normal">
                      <legend class="label">Wireguard Provider</legend>
                    </div>
                    <div class="field-body">
                      <div class="field">
                        <div class="control select is-fullwidth">
                          <select id="wireguard-provider" v-model="wireguard.config.serverName"
                            @change="setModified('wireguard')">
                            <option v-for="server in wireguard.servers" :key="server.name" :value="server.name"
                              :selected="server.name === wireguard.config.serverName">
                              {{ server.name }}
                            </option>
                          </select>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              </form>
              <div class="mt-4 buttons">
                <button class="button is-success mx-auto" @click="saveConfig" :disabled="!isModified">Save</button>
              </div>
            </div>
          </div>
          <div class="container column">
            <div class="container box" style="height: 100%;">
              <div class="level">
                <div class="level-left">
                  <h3 class="title is-4">IP Info</h3>
                </div>
                <!-- <div class="level-right">
                  <div class="buttons">
                    <div class="tooltip">
                      <button v-if="!openvpn.status" class="button is-small is-success" @click="vpnCommand('start')">
                        <span class="icon">
                          <i class="fas fa-play"></i>
                        </span>
                      </button>
                      <span class="tooltip-text">Start OpenVPN Server</span>
                    </div>
                    <div class="tooltip">
                      <button v-if="openvpn.status" class="button is-small is-warning" @click="vpnCommand('stop')">
                        <span class="icon">
                          <i class="fas fa-stop"></i>
                        </span>
                      </button>
                      <span class="tooltip-text">Stop OpenVPN Server</span>
                    </div>
                    <div class="tooltip">
                      <button v-if="openvpn.status" class="button is-small is-info" @click="vpnCommand('restart')">
                        <span class="icon">
                          <i class="fas fa-redo"></i>
                        </span>
                      </button>
                      <span class="tooltip-text">Restart OpenVPN Server</span>
                    </div>
                    <div class="tooltip">
                      <button v-if="openvpn.status" class="button is-small is-light" @click="refreshInfo">
                        <span class="icon">
                          <i class="fas fa-sync-alt"></i>
                        </span>
                      </button>
                      <span class="tooltip-text">Refresh Status</span>
                    </div>
                  </div>
                </div> -->
              </div>
              <div v-if="ipInfo">
                <div class="container">
                  <div class="columns">
                    <div class="column is-3 has-text-weight-bold">IP Address:</div>
                    <div class="column" id="ip">{{ ipInfo.ip }}</div>
                  </div>

                  <div class="columns">
                    <div class="column is-3 has-text-weight-bold">Provider:</div>
                    <div class="column" id="org">{{ ipInfo.org }}</div>
                  </div>

                  <div class="columns">
                    <div class="column is-3 has-text-weight-bold">Location:</div>
                    <div class="column" id="location">{{ ipInfo.city }}, {{ ipInfo.country }}</div>
                  </div>

                  <div class="columns">
                    <div class="column is-3 has-text-weight-bold">Timezone:</div>
                    <div class="column" id="timezone">{{ ipInfo.timezone }}</div>
                  </div>
                </div>
                <!-- Map Display -->
                <location-map class="mt-4" v-model:latitude="ipInfo.loc.split(',')[0]"
                  v-model:longitude="ipInfo.loc.split(',')[1]" v-model:city="ipInfo.city">
                </location-map>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- OpenVPN Servers Tab -->
      <div v-if="currentTab === 'openvpn'" class="box">
        <openvpn-config v-model:servers="openvpn.servers">
        </openvpn-config>
      </div>

      <!-- Wireguard Servers Tab -->
      <div v-if="currentTab === 'wireguard'" class="box">
        <wireguard-config v-model:servers="wireguard.servers">
        </wireguard-config>
      </div>
    </div>
  </section>
  <!-- Footer Section -->
  <footer class="footer">
    <div class="content has-text-centered">
      <p>Follow the project on:</p>
      <div class="buttons is-centered are-medium">
        <!-- GitHub Button -->
        <a href="https://github.com/vm75/vpn-sandbox" target="_blank" class="button is-dark">
          <span class="icon">
            <img src="assets/github.svg" alt="GitHub" style="width: 1em; height: 1em; filter: invert(1);">
          </span>
          <span>GitHub</span>
        </a>

        <!-- Docker Hub Button -->
        <a href="https://hub.docker.com/repository/docker/vm75/vpn-sandbox" target="_blank" class="button is-info">
          <span class="icon">
            <img src="assets/docker.svg" alt="Docker" style="width: 1em; height: 1em;">
          </span>
          <span>Docker Hub</span>
        </a>
      </div>
      <!-- Attribution Link -->
      <a href="https://www.flaticon.com/free-icons/vpn" title="vpn icons" target="_blank" class="attribution-link">
        Vpn icons created by Ranah Pixel Studio - Flaticon
      </a>
    </div>
  </footer>
</template>

<script>
export default {
  data() {
    return {
      currentTab: 'config',
      ipInfo: null,
      openvpn: {
        status: false,
        modified: false,
        config: {
          enabled: false,
          serverName: '',
          serverEndpoint: '',
          logLevel: 3,
          retryInterval: 3600,
        },
        servers: [],
      },
      wireguard: {
        status: false,
        modified: false,
        config: {
          enabled: false,
          serverName: '',
        },
        servers: [],
      },
      http_proxy: {
        status: false,
        config: {
          enabled: false,
        }
      },
      socks_proxy: {
        status: false,
        config: {
          enabled: false,
        }
      },
      global: {
        modified: false,
        config: {
          vpnType: 'openvpn',
          vpnTypes: ['openvpn', 'wireguard'],
          subnets: [],
          proxyUsername: '',
          proxyPassword: '',
        }
      }
    }
  },
  components: {
    'list-editor': Vue.defineAsyncComponent(() => Component.import('components/core/list-editor')),
    'basic': Vue.defineAsyncComponent(() => Component.import('components/core/basic-input')),
    'inline-list': Vue.defineAsyncComponent(() => Component.import('components/core/inline-list')),
    'location-map': Vue.defineAsyncComponent(() => Component.import('components/core/location-map')),
    'openvpn-config': Vue.defineAsyncComponent(() => Component.import('components/app/openvpn-config')),
    'wireguard-config': Vue.defineAsyncComponent(() => Component.import('components/app/wireguard-config')),
  },
  methods: {
    async reload() {
      var globalConfig = await fetch('/api/config').then(response => response.json());
      Object.assign(this.global.config, {
        vpnType: globalConfig.vpnType || 'openvpn',
        subnets: globalConfig.subnets || [],
      })
      this.global.modified = false;

      var openVPNConfig = await fetch('/api/openvpn/config').then(response => response.json());
      Object.assign(this.openvpn.config, {
        enabled: openVPNConfig.enabled || false,
        serverName: openVPNConfig.serverName || '',
        serverEndpoint: openVPNConfig.serverEndpoint || '',
        logLevel: openVPNConfig.logLevel || 3,
        retryInterval: openVPNConfig.retryInterval || 3600,
      })
      this.openvpn.servers = await fetch('/api/openvpn/servers').then(response => response.json());
      this.openvpn.modified = false;

      var wireguardConfig = await fetch('/api/wireguard/config').then(response => response.json());
      Object.assign(this.wireguard.config, {
        enabled: wireguardConfig.enabled || false,
        serverName: wireguardConfig.serverName || '',
      })
      this.wireguard.servers = await fetch('/api/wireguard/servers').then(response => response.json());
      this.wireguard.modified = false;

      var httpProxyConfig = await fetch('/api/http_proxy/config').then(response => response.json());
      Object.assign(this.http_proxy.config, {
        enabled: httpProxyConfig.enabled || false,
      })
      var socksProxyConfig = await fetch('/api/socks_proxy/config').then(response => response.json());
      Object.assign(this.socks_proxy.config, {
        enabled: socksProxyConfig.enabled || false,
      })
    },
    vpnCommand: function (cmd) {
      fetch(`/api/openvpn/${cmd}`, {
        method: 'POST',
      });
    },
    toggleModule: function (module) {
      if (module === 'vpn') {
        module = 'openvpn';
      }
      var cmd = this[module].config.enabled ? 'enable' : 'disable';
      var now = this[module].config.enabled ? 'start' : 'stop';
      fetch(`/api/${module}/${cmd}?${now}=true`, {
        method: 'POST',
      });
    },
    setModified: function (what) {
      this[what].modified = true;
    },
    saveConfig: function (module) {
      if (this.openvpn.modified) {
        fetch(`/api/openvpn/config/save`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(this.openvpn.config)
        });
      }
      if (this.global.modified) {
        fetch(`/api/config/save`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(this.global.config)
        });
      }
    },
    refreshInfo: function () {
      try {
        fetch('/api/openvpn/status').then(response => response.json()).then(data => {
          this.vpnStatus = data;
          this.ipInfo = data.info;
        });
      } catch (error) {
        // console.log(error);
      }
    },
  },
  computed: {
    endpoints: function () {
      for (const server of this.openvpn.servers) {
        if (server.name === this.openvpn.config.serverName) {
          return server.endpoints;
        }
      }
      return [];
    },
    isModified: function () {
      return this.openvpn.modified || this.global.modified;
    }
  },
  mounted() {
    this.reload();
    this.refreshInfo();
    setInterval(() => {
      this.refreshInfo();
    }, 60000);
  }
}
</script>

<style>
.tooltip {
  position: relative;
  display: inline-block;
}

/* Tooltip text */
.tooltip .tooltip-text {
  visibility: hidden;
  width: auto;
  background-color: black;
  color: white;
  text-align: center;
  padding: 8px;
  border-radius: 4px;
  opacity: 0.8;
  /* Set opacity */
  position: absolute;
  bottom: -35px;
  /* Position below the button */
  left: 50%;
  transform: translateX(-50%);
  white-space: nowrap;
  /* Prevent line break */
  z-index: 1;
  font-size: 14px;
  pointer-events: none;
}

/* Show the tooltip when hovering over the tooltip container */
.tooltip:hover .tooltip-text {
  visibility: visible;
}

/* Custom style for attribution */
.attribution-link {
  font-size: 0.75rem;
  /* Smaller font size */
  position: absolute;
  bottom: 20px;
  right: 30px;
  color: #555;
  /* Light gray color */
}
</style>