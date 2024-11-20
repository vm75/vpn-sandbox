<template>
  <div id="modal"></div>
  <section>
    <div class="mb-4 is-flex is-justify-content-center is-align-items-center">
      <icon icon="assets/vpn-sandbox.png"></icon>
      <h1 class="title ml-2">VPN Sandbox</h1>
    </div>
  </section>
  <section>
    <div class="container">

      <!-- Tabs -->
      <div class="tabs is-boxed">
        <ul>
          <li :class="{ 'is-active': currentTab === 'config' }">
            <a @click="currentTab = 'config'">Sandbox Config</a>
          </li>
          <li :class="{ 'is-active': currentTab === 'OpenVPN' }">
            <a @click="currentTab = 'OpenVPN'">OpenVPN Servers</a>
          </li>
          <li :class="{ 'is-active': currentTab === 'Wireguard' }">
            <a @click="currentTab = 'Wireguard'">Wireguard Servers</a>
          </li>
          <li :class="{ 'is-active': currentTab === 'Files' }">
            <a @click="currentTab = 'Files'">Runtime Files</a>
          </li>
        </ul>
      </div>

      <!-- Sandbox Config -->
      <div v-if="currentTab === 'config'">
        <div class="columns">
          <div class="container column">
            <div class="container box" style="height: 100%;">
              <form>
                <app-status v-if="global.vpnType === 'OpenVPN'" name="openvpn" displayName="VPN"
                  v-model:enabled="openvpn.config.enabled" v-model:running="openvpn.running"
                  @toggleModule="toggleModule">
                </app-status>
                <app-status v-if="global.vpnType === 'Wireguard'" name="wireguard" displayName="VPN"
                  v-model:enabled="wireguard.config.enabled" v-model:running="wireguard.running"
                  @toggleModule="toggleModule">
                </app-status>
                <app-status name="http_proxy" displayName="Http Proxy" v-model:enabled="http_proxy.config.enabled"
                  v-model:running="http_proxy.running" @toggleModule="toggleModule">
                </app-status>
                <app-status name="socks_proxy" displayName="Socks Proxy" v-model:enabled="socks_proxy.config.enabled"
                  v-model:running="socks_proxy.running" @toggleModule="toggleModule">
                </app-status>
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
                        type="subnet" @update:entries="setModified">
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
                        <select id="vpn-type" v-model="global.config.vpnType" @change="setModified">
                          <option v-for="vpnType in global.config.vpnTypes" :key="vpnType" :value="vpnType"
                            :selected="vpnType === 'OpenVPN'">
                            {{ vpnType }}
                          </option>
                        </select>
                      </div>
                    </div>
                  </div>
                </div>
                <div v-if="global.config.vpnType === 'OpenVPN'">
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
                          <select id="openvpn-provider" v-model="openvpn.config.serverName" @change="setModified">
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
                        <select id="openvpn-endpoint" v-model="openvpn.config.serverEndpoint" @change="setModified">
                          <option v-for="endpoint in endpoints" :key="endpoint.name" :value="endpoint.name"
                            :selected="endpoint.name === openvpn.config.serverEndpoint">
                            {{ endpoint.name }}
                          </option>
                        </select>
                      </div>
                    </div>
                  </div>
                </div>
                <div v-if="global.config.vpnType === 'Wireguard'">
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
                          <select id="wireguard-provider" v-model="wireguard.config.serverName" @change="setModified">
                            <option v-for="server in wireguard.servers" :key="server.name" :value="server.name"
                              :selected="server.name === wireguard.config.serverName">
                              {{ server.name }}
                            </option>
                          </select>
                        </div>
                      </div>
                    </div>
                  </div>
                  <div v-if="serverHasParams" class="field is-horizontal">
                    <div class="field-label is-normal">
                      <legend class="label">Server Endpoint</legend>
                    </div>
                    <div class="field-body">
                      <div class="field control select is-fullwidth">
                        <select id="wireguard-endpoint" v-model="wireguard.config.serverEndpoint" @change="setModified">
                          <option v-for="endpoint in endpoints" :key="endpoint.name" :value="endpoint.name"
                            :selected="endpoint.name === wireguard.config.serverEndpoint">
                            {{ endpoint.name }}
                          </option>
                        </select>
                      </div>
                    </div>
                  </div>
                </div>
              </form>
              <div class="mt-4 buttons">
                <button class="button is-info mx-auto" @click="refreshInfo(false)">Reset</button>
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
                <div class="level-right">
                  <div class="buttons">
                    <div class="tooltip">
                      <button class="button is-small is-light" @click="refreshInfo(true)">
                        <span class="icon">
                          <i class="fas fa-sync-alt"></i>
                        </span>
                      </button>
                      <span class="tooltip-text">Refresh Status</span>
                    </div>
                  </div>
                </div>
              </div>
              <div v-if="ipInfo && Object.keys(ipInfo).length > 0">
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
                    <div class="column" id="location">
                      {{ ipInfo.city }}, {{ ipInfo.region }}, {{ ipInfo.country }}, {{ ipInfo.postal }}
                    </div>
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
      <div v-if="currentTab === 'OpenVPN'" class="box">
        <vpn-config vpnType="OpenVPN" v-model:servers="openvpn.servers" />
      </div>

      <!-- Wireguard Servers Tab -->
      <div v-if="currentTab === 'Wireguard'" class="box">
        <vpn-config vpnType="Wireguard" v-model:servers="wireguard.servers" />
      </div>

      <!-- File Browser Tab -->
      <div v-if="currentTab === 'Files'" class="box">
        <file-browser-viewer filesEndpoint="/api/files" fileEndpoint="/api/file" />
      </div>
    </div>
  </section>
  <!-- Footer Section -->
  <section>
    <div class="mt-4 content has-text-centered">
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
  </section>
</template>

<script>

const REFRESH_TIME = 3000;

// Main App Component
export default {
  data() {
    return {
      currentTab: 'config',
      global: {
        modified: false,
        vpnType: 'OpenVPN',
        config: {
          vpnType: 'OpenVPN',
          vpnTypes: ['OpenVPN', 'Wireguard'],
          subnets: [],
          proxyUsername: '',
          proxyPassword: '',
        }
      },
      openvpn: {
        modified: false,
        running: false,
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
        running: false,
        modified: false,
        config: {
          enabled: false,
          serverName: '',
          serverEndpoint: '',
        },
        servers: [],
      },
      http_proxy: {
        running: false,
        config: {
          enabled: false,
        }
      },
      socks_proxy: {
        running: false,
        config: {
          enabled: false,
        }
      },
      ipInfo: null,
    }
  },
  components: {
    'list-editor': Vue.defineAsyncComponent(() => ComponentLoader.import('core/list-editor')),
    'basic': Vue.defineAsyncComponent(() => ComponentLoader.import('core/basic-input')),
    'inline-list': Vue.defineAsyncComponent(() => ComponentLoader.import('core/inline-list')),
    'location-map': Vue.defineAsyncComponent(() => ComponentLoader.import('core/location-map')),
    'vpn-config': Vue.defineAsyncComponent(() => ComponentLoader.import('app/vpn-config')),
    'app-status': Vue.defineAsyncComponent(() => ComponentLoader.import('app/app-status')),
    'icon': Vue.defineAsyncComponent(() => ComponentLoader.import('core/icon')),
    'file-browser-viewer': Vue.defineAsyncComponent(() => ComponentLoader.import('core/file-browser-viewer')),
  },
  methods: {
    async refreshInfo(force) {
      var status = await fetch(`/api/status?force=${force}`).then(response => response.json());

      // console.log(status);

      this.global.config = status.global.config;
      this.global.vpnType = status.global.config.vpnType;
      this.global.modified = false;

      this.openvpn.running = status.openvpn.running;
      var openVPNConfig = status.openvpn.config;
      this.openvpn.config = openVPNConfig;
      this.openvpn["servers"] = openVPNConfig.servers || [];
      this.openvpn.modified = false;

      this.wireguard.running = status.wireguard.running;
      var wireguardConfig = status.wireguard.config;
      this.wireguard.config = wireguardConfig;
      this.wireguard["servers"] = wireguardConfig.servers || [];
      this.wireguard.modified = false;

      this.http_proxy.running = status.http_proxy.running;
      this.http_proxy.config = status.http_proxy.config;

      this.socks_proxy.running = status.socks_proxy.running;
      this.socks_proxy.config = status.socks_proxy.config;

      this.ipInfo = status.ipInfo;
    },
    toggleModule: function (module) {
      this[module].config.enabled = !this[module].config.enabled;
      var cmd = this[module].config.enabled ? 'enable' : 'disable';
      var now = this[module].config.enabled ? 'start' : 'stop';
      fetch(`/api/${module}/${cmd}?${now}=true`, {
        method: 'POST',
      }).then(() => {
        setTimeout(() => {
          this.refreshInfo(false);
        }, REFRESH_TIME);
      });
    },
    setModified: function (event) {
      switch (event.target.id) {
        case 'lan-subnets':
          this.global.modified = true;
          break;
        case 'vpn-type':
          this.global.modified = true;
          break;
        case 'openvpn-provider':
          this.openvpn.config.serverEndpoint = '';
          this.openvpn.modified = true;
          break;
        case 'openvpn-endpoint':
          this.openvpn.modified = true;
          break;
        case 'wireguard-provider':
          this.wireguard.config.serverEndpoint = '';
          this.wireguard.modified = true;
          break;
        case 'wireguard-endpoint':
          this.wireguard.modified = true;
          break;
      }
    },
    saveConfig: async function () {
      var configTypes = ['global', 'openvpn', 'wireguard'];

      for (var configType of configTypes) {
        if (this[configType].modified) {
          await fetch(`/api/${configType === 'global' ? '' : (configType + '/')}config/save`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json'
            },
            body: JSON.stringify(this[configType].config)
          });
          this[configType].modified = false;
        }
      }
      if (this.global.vpnType !== this.global.config.vpnType) {
        var orgType = this.global.vpnType.toLowerCase();
        var newType = this.global.config.vpnType.toLowerCase();
        var vpnEnabled = this[orgType].config.enabled;
        if (vpnEnabled) {
          await fetch(`/api/${orgType}/disable?stop=true`, {
            method: 'POST',
          });
          await fetch(`/api/${newType}/enable?start=true`, {
            method: 'POST',
          });
        }
      }
      setTimeout(() => {
        this.refreshInfo(false);
      }, REFRESH_TIME);
    },
  },
  computed: {
    vpnEnabled: function () {
      if (this.global.config.vpnType === 'OpenVPN') {
        return this.openvpn.config.enabled;
      } else if (this.global.config.vpnType === 'Wireguard') {
        return this.wireguard.config.enabled;
      }
      return false;
    },
    vpnRunning: function () {
      if (this.global.config.vpnType === 'OpenVPN') {
        return this.openvpn.running;
      } else if (this.global.config.vpnType === 'Wireguard') {
        return this.wireguard.running;
      }
      return false;
    },
    endpoints: function () {
      for (const server of this.openvpn.servers) {
        if (server.name === this.openvpn.config.serverName) {
          return server.endpoints;
        }
      }
      return [];
    },
    serverHasParams: function () {
      var vpnModule = this.global.config.vpnType.toLowerCase();
      var server = this[vpnModule].servers.find(server => server.name === this[vpnModule].config.serverName);
      return server && server.hasParams;
    },
    isModified: function () {
      var vpnModule = this.global.config.vpnType.toLowerCase();
      if (!this[vpnModule].config.serverName) {
        return false;
      }
      if (this.serverHasParams && !this[vpnModule].config.serverEndpoint) {
        return false;
      }
      return this.global.modified || this[vpnModule].modified || this.wireguard.modified;
    }
  },
  mounted() {
    this.refreshInfo(false);
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
  bottom: 10px;
  right: 20px;
  color: #555;
  /* Light gray color */
}
</style>