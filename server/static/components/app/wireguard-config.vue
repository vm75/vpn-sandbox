<template>
  <div id="wireguard-config-modal"></div>
  <list-editor :name="'Wireguard Servers'" :list="serverList" :editItem="editServer" :addItem="editServer"
    :removeItem="deleteServer" :displayString="serversDisplayString">
  </list-editor>
</template>

<script>
// Wireguard Config Component
export default {
  name: 'wireguard-config',
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
        elementId: "wireguard-config-modal",
        name: 'edit-wireguard',
        source: 'app/edit-wireguard',
        data: {
          server: index !== undefined ? this.serverList[index] : {},
          showOnLoad: true
        },
        methods: {
          save: (data) => {
            fetch('/api/wireguard/servers/save', {
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
      fetch(`/api/wireguard/servers/delete/${this.serverList[index].name}`, {
        method: 'DELETE'
      }).then(() => {
        this.refresh();
      });
    },
    serversDisplayString: function (server) {
      return server.name;
    },
    refresh: function () {
      fetch('/api/wireguard/servers').then(response => response.json()).then(data => {
        this.serverList = data;
        this.$emit('update:servers', this.serverList);
      })
    }
  },
}
</script>
