<template>
  <div v-if="isVisible">
    <div class="modal is-active">
      <div class="modal-background"></div>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ server.name ? 'Edit OpenVPN Provider' : 'New OpenVPN Provider' }}</p>
          <button class="delete" aria-label="close" @click="cancel"></button>
        </header>
        <section class="modal-card-body">
          <div class="field">
            <legend class="label">OpenVPN Provider</legend>
            <div class="control">
              <input id="openvpn-provider" class="input" v-model="server.name" placeholder="OpenVPN Provider"
                :disabled="!nameIsEditable" />
            </div>
          </div>
          <div class="field">
            <legend class="label">Username</legend>
            <div class="control">
              <input id="openvpn-username" class="input" v-model="server.username" placeholder="Username" />
            </div>
          </div>
          <div class="field">
            <legend class="label">Password</legend>
            <div class="control">
              <input id="openvpn-password" class="input" type="password" v-model="server.password"
                placeholder="Password" />
            </div>
          </div>
          <div class="field">
            <legend class="label">Endpoints</legend>
            <div class="control vue-bulma-input">
              <table class="table is-fullwidth is-striped">
                <thead>
                  <tr>
                    <th>Actions</th>
                    <th>Endpoint Name</th>
                    <th v-for="(variable, index) in variables" :key="'var' + index">{{ variable }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(endpoint, index) in server.endpoints" :key="'endpoint' + index">
                    <td>
                      <button class="button is-small is-danger" @click="deleteEndpoint(index)">🗑</button>
                    </td>
                    <td>
                      <input class="input" v-model="endpoint.name" placeholder="name" />
                    </td>
                    <td v-for="(variable, vindex) in variables" :key="'varinput' + vindex">
                      <input class="input" v-model="server.endpoints[index][variable]" :placeholder="variable" />
                    </td>
                  </tr>
                  <tr>
                    <td>
                      <button class="button is-small is-info" @click="newEndpoint()">➕</button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
          <div class="field">
            <legend class="label">ovpn Template</legend>
            <div class="control">
              <textarea id="openvpn-template" class="textarea" v-model="server.template" placeholder="ovpn content"
                style="white-space: pre; overflow-x: auto; font-family: 'Courier New', Courier, monospace;"></textarea>
            </div>
          </div>
        </section>
        <footer class="modal-card-foot">
          <button class="button mx-auto" @click="cancel">Cancel</button>
          <button class="button is-success mx-auto" @click="save">Save</button>
        </footer>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: "edit-openvpn",
  props: {
    server: {
      type: Object,
      required: true,
      default: () => {
        return {
          name: '',
          username: '',
          password: '',
          endpoints: [],
          template: '',
        }
      },
    },
    showOnLoad: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    this.server.endpoints = this.server.endpoints || [];
    this.server.template = this.server.template || '';
    return {
      isVisible: this.showOnLoad || false,
      nameIsEditable: true,
    }
  },
  methods: {
    show(name, content) {
      this.isVisible = true;
      this.server.name = name;
      this.server.content = content;
    },
    cancel() {
      this.isVisible = false;
      this.$emit('cancel');
    },
    save() {
      this.isVisible = false;
      this.$emit('save', this.server);
    },
    newEndpoint() {
      // allow only one empty endpoint name
      for (const endpoint in this.server.endpoints) {
        if (endpoint.name === '') {
          return;
        }
      }

      const endpoint = { name: '' };
      for (const variable of this.variables) {
        endpoint[variable] = '';
      }
      this.server.endpoints.push(endpoint);
    },
    deleteEndpoint(index) {
      this.server.endpoints.splice(index, 1);
    }
  },
  computed: {
    variables() {
      // extract all variables from the template content which are enclosed in {{}}
      const re = /\{\{\s*(.*?)\s*\}\}/g;
      const match = this.server.template.match(re);
      if (match) {
        return match.map((m) => m.substring(2, m.length - 2));
      }
      return [];
    }
  },
  mounted() {
    this.nameIsEditable = !this.server.name;
  }
}
</script>

<style>
.vue-bulma-input {
  border: 1px solid #dbdbdb;
  border-radius: 5px;
  padding: 0.5rem;
}

.vue-bulma-input:focus-within {
  border-color: hsl(var(--bulma-input-focus-h), var(--bulma-input-focus-s), var(--bulma-input-focus-l));
  box-shadow: var(--bulma-input-focus-shadow-size) hsla(var(--bulma-input-focus-h), var(--bulma-input-focus-s), var(--bulma-input-focus-l), var(--bulma-input-focus-shadow-alpha));
  outline: none;
}
</style>
