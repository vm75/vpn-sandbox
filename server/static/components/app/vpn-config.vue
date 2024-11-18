<template>
  <div id="vpn-config-modal"></div>
  <list-editor :name="vpnType + ' Servers'" :list="serverList" :editItem="editServer" :addItem="editServer"
    :removeItem="deleteServer" :displayString="serversDisplayString">
  </list-editor>
</template>

<script>
// VPN Config Component
export default {
  name: 'vpn-config',
  props: {
    vpnType: {
      type: String,
      required: true
    },
    servers: {
      type: Array,
      required: true
    }
  },
  data() {
    return {
      serverList: this.servers,
      serverModule: this.vpnType.toLowerCase()
    }
  },
  components: {
    'list-editor': Vue.defineAsyncComponent(() => ComponentLoader.import('core/list-editor')),
  },
  methods: {
    editServer: function (index) {
      ComponentLoader.inject({
        elementId: "vpn-config-modal",
        name: 'vpn-config-edit',
        source: 'app/vpn-config-edit',
        data: {
          vpnType: this.vpnType,
          server: index !== undefined ? this.serverList[index] : {},
          showOnLoad: true
        },
        methods: {
          save: (data) => {
            if (this.vpnType === 'Wireguard') {
              delete data.username;
              delete data.password;
            }
            fetch(`/api/${this.serverModule}/servers/save`, {
              method: 'POST',
              headers: {
                'Content-Type': 'application/json'
              },
              body: JSON.stringify(data)
            }).then(() => {
              this.refresh();
            });
          }
        }
      });
    },
    deleteServer: function (index) {
      fetch(`/api/${this.serverModule}/servers/delete/${this.serverList[index].name}`, {
        method: 'DELETE'
      }).then(() => {
        this.refresh();
      });
    },
    serversDisplayString: function (server) {
      return server.name;
    },
    refresh: function () {
      fetch(`/api/${this.serverModule}/servers`).then(response => response.json()).then(data => {
        this.serverList = data;
        this.$emit('update:servers', this.serverList);
      })
    }
  },
}
</script>
