<template>
  <div id="openvpn-config-modal"></div>
  <list-editor :name="'OpenVPN Servers'" :list="serverList" :editItem="editServer" :addItem="editServer"
    :removeItem="deleteServer" :displayString="serversDisplayString">
  </list-editor>
</template>

<script>
// OpenVPN Config Component
export default {
  name: 'openvpn-config',
  props: {
    servers: {
      type: Array,
      required: true
    }
  },
  data() {
    return {
      serverList: this.servers,
    }
  },
  components: {
    'list-editor': Vue.defineAsyncComponent(() => ComponentLoader.import('core/list-editor')),
  },
  methods: {
    editServer: function (index) {
      ComponentLoader.inject({
        elementId: "openvpn-config-modal",
        name: 'edit-openvpn',
        source: 'app/edit-openvpn',
        data: {
          server: index !== undefined ? this.serverList[index] : {},
          showOnLoad: true
        },
        methods: {
          save: (data) => {
            fetch('/api/openvpn/servers/save', {
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
      fetch(`/api/openvpn/servers/delete/${this.serverList[index].name}`, {
        method: 'DELETE'
      }).then(() => {
        this.refresh();
      });
    },
    serversDisplayString: function (server) {
      return server.name;
    },
    refresh: function () {
      fetch('/api/openvpn/servers').then(response => response.json()).then(data => {
        this.serverList = data;
        this.$emit('update:servers', this.serverList);
      })
    }
  },
}
</script>
