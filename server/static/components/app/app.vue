<template>
  <div id="modal"></div>
  <section>
    <div class="mb-4 is-flex is-justify-content-center is-align-items-center">
      <icon icon="assets/vpn-sandbox.png"></icon>
      <div class="ml-2 is-flex is-align-items-baseline">
        <h1 class="title mb-1">VPN Sandbox</h1>
        <small v-if="version" class="ml-2 has-text-grey">v{{ version }}</small>
      </div>
    </div>
  </section>
  <section>
    <div class="container">

      <!-- Tabs -->
      <div class="tabs is-boxed">
        <ul>
          <li :class="{ 'is-active': currentTab === 'config' }">
            <a @click="currentTab = 'config'">Dashboard</a>
          </li>
          <li :class="{ 'is-active': currentTab === 'OpenVPN' }">
            <a @click="currentTab = 'OpenVPN'">OpenVPN Servers</a>
          </li>
          <li :class="{ 'is-active': currentTab === 'Wireguard' }">
            <a @click="currentTab = 'Wireguard'">Wireguard Servers</a>
          </li>
          <li :class="{ 'is-active': currentTab === 'Apps' }">
            <a @click="currentTab = 'Apps'">Apps</a>
          </li>
          <li :class="{ 'is-active': currentTab === 'Files' }">
            <a @click="currentTab = 'Files'">Runtime Files</a>
          </li>
        </ul>
      </div>

      <!-- Dashboard -->
      <div v-if="currentTab === 'config'">
        <div class="columns">
          <div class="container column">
            <div class="container box is-flex is-flex-direction-column" style="height: 100%;">
              <!-- Header -->
              <div class="level mb-5" style="border-bottom: 1px solid #eee; padding-bottom: 10px;">
                <div class="level-left">
                  <h3 class="title is-4 mb-0">
                    <span class="icon-text is-align-items-center">
                      <span class="icon has-text-primary mr-2"><i class="fas fa-sliders-h"></i></span>
                      <span>Configuration</span>
                    </span>
                  </h3>
                </div>
              </div>

              <!-- Form Section -->
              <form class="is-flex-grow-1">
                <div class="box has-background-light p-4 mb-5" style="border-radius: 8px; box-shadow: none; border: 1px solid #eaeaea;">
                  <h4 class="title is-6 has-text-grey-dark mb-4">
                    <span class="icon mr-2"><i class="fas fa-power-off"></i></span>Service Status
                  </h4>
                  <app-status v-if="global.vpnType === 'OpenVPN'" name="openvpn" displayName="VPN"
                    v-model:enabled="openvpn.config.enabled" v-model:running="openvpn.running"
                    @toggleModule="toggleModule">
                  </app-status>
                  <app-status v-if="global.vpnType === 'Wireguard'" name="wireguard" displayName="VPN"
                    v-model:enabled="wireguard.config.enabled" v-model:running="wireguard.running"
                    @toggleModule="toggleModule">
                  </app-status>
                  <app-status name="http_proxy" displayName="HTTP Proxy" v-model:enabled="http_proxy.config.enabled"
                    v-model:running="http_proxy.running" @toggleModule="toggleModule">
                  </app-status>
                  <app-status name="socks_proxy" displayName="SOCKS Proxy" v-model:enabled="socks_proxy.config.enabled"
                    v-model:running="socks_proxy.running" @toggleModule="toggleModule">
                  </app-status>
                </div>

                <div class="box has-background-light p-4 mb-5" style="border-radius: 8px; box-shadow: none; border: 1px solid #eaeaea;">
                  <h4 class="title is-6 has-text-grey-dark mb-4">
                    <span class="icon mr-2"><i class="fas fa-cogs"></i></span>Common Config
                  </h4>
                  <div class="field is-horizontal mb-4">
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
                </div>

                <div v-if="global.config.vpnType === 'OpenVPN'" class="box has-background-light p-4 mb-5" style="border-radius: 8px; box-shadow: none; border: 1px solid #eaeaea;">
                  <h4 class="title is-6 has-text-grey-dark mb-4">
                    <span class="icon mr-2"><i class="fas fa-shield-alt"></i></span>OpenVPN Config
                  </h4>
                  <div class="field is-horizontal mb-4">
                    <div class="field-label is-normal">
                      <legend class="label">Provider</legend>
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
                      <legend class="label">Endpoint</legend>
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

                <div v-if="global.config.vpnType === 'Wireguard'" class="box has-background-light p-4 mb-5" style="border-radius: 8px; box-shadow: none; border: 1px solid #eaeaea;">
                  <h4 class="title is-6 has-text-grey-dark mb-4">
                    <span class="icon mr-2"><i class="fas fa-shield-alt"></i></span>WireGuard Config
                  </h4>
                  <div class="field is-horizontal mb-4">
                    <div class="field-label is-normal">
                      <legend class="label">Provider</legend>
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
                      <legend class="label">Endpoint</legend>
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

              <!-- Actions -->
              <div class="mt-auto pt-4 has-text-centered" style="border-top: 1px solid #eee;">
                <div class="buttons is-centered">
                  <button class="button is-light is-rounded px-5" @click="forceRefresh">
                    <span class="icon"><i class="fas fa-undo"></i></span>
                    <span>Reset</span>
                  </button>
                  <button class="button is-primary is-rounded px-5" @click="saveConfig" :disabled="!isModified">
                    <span class="icon"><i class="fas fa-save"></i></span>
                    <span>Save Changes</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
          <div class="container column">
            <div class="container box" style="height: 100%;">
              <div class="level mb-4" style="border-bottom: 1px solid #eee; padding-bottom: 10px;">
                <div class="level-left">
                  <h3 class="title is-4 mb-0">
                    <span class="icon-text is-align-items-center">
                      <span class="icon has-text-info mr-2"><i class="fas fa-globe"></i></span>
                      <span>IP Info</span>
                    </span>
                  </h3>
                  <span v-if="ipInfo" class="tag is-rounded ml-3" :class="ipInfo.stale ? 'is-warning is-light' : 'is-success is-light'">
                    {{ ipInfo.stale ? 'Stale' : 'Fresh' }}
                  </span>
                </div>
                <div class="level-right">
                  <div class="buttons">
                    <div class="tooltip">
                      <button class="button is-small is-rounded is-light is-info is-outlined" @click="forceRefresh">
                        <span class="icon">
                          <i class="fas fa-sync-alt"></i>
                        </span>
                      </button>
                      <span class="tooltip-text">Refresh Status</span>
                    </div>
                  </div>
                </div>
              </div>
              <div v-if="ipInfo">
                <div class="is-flex is-justify-content-space-between content is-small has-text-grey mb-4 p-3 has-background-light" style="border-radius: 8px;">
                  <div>
                    <span class="icon mr-1"><i class="fas fa-play-circle"></i></span>
                    Last executed:
                    <span class="has-text-weight-medium">{{ formatTimestamp(ipInfo.executedAt) }}</span>
                  </div>
                  <div>
                    <span class="icon mr-1"><i class="fas fa-bolt"></i></span>
                    Last trigger:
                    <span class="has-text-weight-medium">
                      {{ formatIpInfoEvent(ipInfo.event) }} at {{ formatTimestamp(ipInfo.eventAt) }}
                    </span>
                  </div>
                </div>
              </div>
              <div v-if="ipInfo && ipInfo.output && Object.keys(ipInfo.output).length > 0">
                <div class="columns is-multiline">
                  <!-- IP Address -->
                  <div class="column is-6">
                    <div class="box p-3" style="box-shadow: 0 2px 5px rgba(0,0,0,0.05); height: 100%;">
                      <div class="is-flex is-align-items-center mb-2">
                        <span class="icon has-text-info mr-2"><i class="fas fa-network-wired"></i></span>
                        <span class="is-uppercase has-text-grey-light is-size-7 has-text-weight-bold">IP Address</span>
                      </div>
                      <div class="is-size-5 has-text-weight-semibold" id="ip">{{ ipInfo.output.ip }}</div>
                    </div>
                  </div>

                  <!-- Provider -->
                  <div class="column is-6">
                    <div class="box p-3" style="box-shadow: 0 2px 5px rgba(0,0,0,0.05); height: 100%;">
                      <div class="is-flex is-align-items-center mb-2">
                        <span class="icon has-text-primary mr-2"><i class="fas fa-server"></i></span>
                        <span class="is-uppercase has-text-grey-light is-size-7 has-text-weight-bold">Provider</span>
                      </div>
                      <div class="is-size-6" id="org">{{ ipInfo.output.org }}</div>
                    </div>
                  </div>

                  <!-- Location -->
                  <div class="column is-6">
                    <div class="box p-3" style="box-shadow: 0 2px 5px rgba(0,0,0,0.05); height: 100%;">
                      <div class="is-flex is-align-items-center mb-2">
                        <span class="icon has-text-danger mr-2"><i class="fas fa-map-marker-alt"></i></span>
                        <span class="is-uppercase has-text-grey-light is-size-7 has-text-weight-bold">Location</span>
                      </div>
                      <div class="is-size-6" id="location">
                        {{ ipInfo.output.city }}, {{ ipInfo.output.region }}, {{ ipInfo.output.country }}, {{ ipInfo.output.postal }}
                      </div>
                    </div>
                  </div>

                  <!-- Timezone -->
                  <div class="column is-6">
                    <div class="box p-3" style="box-shadow: 0 2px 5px rgba(0,0,0,0.05); height: 100%;">
                      <div class="is-flex is-align-items-center mb-2">
                        <span class="icon has-text-warning mr-2"><i class="fas fa-clock"></i></span>
                        <span class="is-uppercase has-text-grey-light is-size-7 has-text-weight-bold">Timezone</span>
                      </div>
                      <div class="is-size-6" id="timezone">{{ ipInfo.output.timezone }}</div>
                    </div>
                  </div>

                  <!-- DNS IP Address -->
                  <div class="column is-6" v-if="ipInfo.output.dns">
                    <div class="box p-3" style="box-shadow: 0 2px 5px rgba(0,0,0,0.05); height: 100%;">
                      <div class="is-flex is-align-items-center mb-2">
                        <span class="icon has-text-info mr-2"><i class="fas fa-network-wired"></i></span>
                        <span class="is-uppercase has-text-grey-light is-size-7 has-text-weight-bold">DNS IP</span>
                      </div>
                      <div class="is-size-6" id="dns-ip">{{ ipInfo.output.dns.ip }}</div>
                    </div>
                  </div>

                  <!-- DNS Provider -->
                  <div class="column is-6" v-if="ipInfo.output.dns">
                    <div class="box p-3" style="box-shadow: 0 2px 5px rgba(0,0,0,0.05); height: 100%;">
                      <div class="is-flex is-align-items-center mb-2">
                        <span class="icon has-text-primary mr-2"><i class="fas fa-server"></i></span>
                        <span class="is-uppercase has-text-grey-light is-size-7 has-text-weight-bold">DNS Provider</span>
                      </div>
                      <div class="is-size-6" id="dns-geo">{{ ipInfo.output.dns.geo }}</div>
                    </div>
                  </div>
                </div>
                <!-- Map Display -->
                <div class="mt-4" style="border-radius: 12px; overflow: hidden; box-shadow: 0 4px 15px rgba(0,0,0,0.1); border: 1px solid #eaeaea;">
                  <location-map v-if="ipInfo.output.loc"
                    v-model:latitude="ipInfo.output.loc.split(',')[0]"
                    v-model:longitude="ipInfo.output.loc.split(',')[1]" v-model:city="ipInfo.output.city">
                  </location-map>
                </div>
              </div>
              <div v-else class="notification is-light">
                Waiting for the first IP info check.
              </div>
            </div>
          </div>
        </div>

        <!-- Apps Panel — shown only when user has configured apps -->
        <div v-if="appsData.length > 0" class="box mt-2">
          <div class="level mb-3">
            <div class="level-left">
              <h3 class="title is-5 mb-0">
                <span class="icon-text">
                  <span class="icon has-text-info"><i class="fas fa-cubes"></i></span>
                  <span>Apps</span>
                </span>
              </h3>
            </div>
            <div class="level-right">
              <a class="is-size-7 has-text-grey" @click="currentTab = 'Apps'" style="cursor:pointer;">
                Manage apps <i class="fas fa-external-link-alt fa-xs"></i>
              </a>
            </div>
          </div>
          <div class="app-toggle-grid">
            <div v-for="userApp in appsData" :key="userApp.name" class="app-toggle-card">
              <div class="app-toggle-icon">
                <span class="icon has-text-info"><i class="fas fa-cube"></i></span>
              </div>
              <div class="app-toggle-name">{{ userApp.name }}</div>
              <label class="toggle-switch app-toggle-switch">
                <input :id="'app-toggle-' + userApp.name" type="checkbox" :checked="userApp.enabled"
                  @change="toggleUserApp(userApp)" />
                <span class="toggle-slider round"></span>
              </label>
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

      <!-- Apps Tab -->
      <div v-if="currentTab === 'Apps'" class="box">
        <apps-manager v-model:apps="appsData" />
      </div>

      <!-- File Browser Tab -->
      <div v-if="currentTab === 'Files'" class="box">
        <file-explorer filesEndpoint="/api/files" fileEndpoint="/api/file" />
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
    </div>
  </section>
</template>

<script>
// Main App Component
export default {
  data() {
    return {
      currentTab: 'config',
      version: '',
      appsData: [],
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
    'file-explorer': Vue.defineAsyncComponent(() => ComponentLoader.import('core/file-explorer')),
    'apps-manager': Vue.defineAsyncComponent(() => ComponentLoader.import('app/apps-manager')),
  },
  methods: {
    formatTimestamp(value) {
      if (!value) {
        return 'Not yet completed';
      }
      return new Date(value).toLocaleString();
    },
    formatIpInfoEvent(value) {
      const labels = {
        'startup': 'Startup',
        'vpn-up': 'Tunnel up',
        'vpn-down': 'Tunnel down',
        'force': 'Manual refresh',
      };
      return labels[value] || value || 'Unknown';
    },
    updateStatus(status) {
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

      if (status.apps && status.apps.config && status.apps.config.apps) {
        this.appsData = status.apps.config.apps;
      }

      this.ipInfo = status.ipInfo;
    },
    async forceRefresh() {
      try {
        const response = await fetch(`/api/force-refresh`);
        if (!response.ok) {
          throw new Error(`IP info refresh failed with status ${response.status}`);
        }
        this.ipInfo = await response.json();
      } catch (error) {
        console.error("Error refreshing IP info:", error);
      }
    },
    toggleModule: function (module) {
      this[module].config.enabled = !this[module].config.enabled;
      var cmd = this[module].config.enabled ? 'enable' : 'disable';
      var now = this[module].config.enabled ? 'start' : 'stop';
      fetch(`/api/${module}/${cmd}?${now}=true`, {
        method: 'POST',
      });
    },
    toggleUserApp: function (userApp) {
      const cmd = userApp.enabled ? 'disable' : 'enable';
      fetch(`/api/apps/${userApp.name}/${cmd}`, { method: 'POST' });
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
      var orgType = this.global.vpnType.toLowerCase();
      var newType = this.global.config.vpnType.toLowerCase();
      var vpnEnabled = this[orgType].config.enabled;
      var vpnTypeChanged = this.global.vpnType !== this.global.config.vpnType;

      if (vpnTypeChanged) {
        await fetch(`/api/${orgType}/disable?stop=true`, {
          method: 'POST',
        });
      }

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

      if (vpnEnabled) {
        await fetch(`/api/${newType}/enable?start=true`, {
          method: 'POST',
        });
      }
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
    fetch("api/version")
      .then(response => response.json())
      .then(data => {
        this.version = data.version;
        document.title = `VPN Sandbox v${data.version}`;
      })
      .catch(error => console.error("Error loading version:", error));

    const eventSource = new EventSource("/api/status");

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        console.log("Received update:", data);
        this.updateStatus(data);
      } catch (error) {
        console.error("Error parsing event data:", error);
      }
    };

    eventSource.onerror = (error) => {
      console.error("SSE connection error:", error);
    };

    this.eventSource = eventSource;
  },
  beforeUnmount() {
    if (this.eventSource) {
      this.eventSource.close();
    }
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

.app-toggle-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
}

.app-toggle-card {
  display: flex;
  align-items: center;
  background-color: #f5f5f5;
  border-radius: 6px;
  padding: 0.75rem 1rem;
  min-width: 200px;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.app-toggle-icon {
  margin-right: 0.75rem;
  font-size: 1.2rem;
}

.app-toggle-name {
  flex-grow: 1;
  font-weight: 500;
  margin-right: 1rem;
}

.app-toggle-switch {
  margin-bottom: 0 !important;
}
</style>
