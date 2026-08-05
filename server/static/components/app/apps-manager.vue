<template>
  <!-- Edit modal -->
  <app-edit :show="showEdit" :app="editingApp" @cancel="closeEdit" @save="onSaveApp"></app-edit>

  <!-- Delete confirm -->
  <confirm-delete :show="deleteTarget !== null" :itemName="deleteTarget || ''" @cancel="deleteTarget = null"
    @confirm="onDeleteApp">
  </confirm-delete>

  <div class="apps-manager">
    <!-- Header row -->
    <div class="level mb-4">
      <div class="level-left">
        <h3 class="title is-5 mb-0">Managed Apps</h3>
      </div>
      <div class="level-right">
        <button id="apps-add-btn" class="button is-success is-small" @click="openAdd">
          <span class="icon"><i class="fas fa-plus"></i></span>
          <span>Add App</span>
        </button>
      </div>
    </div>

    <!-- Empty state -->
    <div v-if="apps.length === 0" class="notification is-light has-text-centered">
      <span class="icon is-large has-text-grey-light"><i class="fas fa-cube fa-2x"></i></span>
      <p class="mt-2 has-text-grey">No apps configured yet.<br>Click <strong>Add App</strong> to define an app.</p>
      <p class="help mt-1">Apps run their <em>up</em> and <em>down</em> commands when the VPN tunnel changes state.</p>
    </div>

    <!-- Apps table -->
    <table v-else class="table is-fullwidth is-striped is-hoverable apps-table">
      <thead>
        <tr>
          <th style="width: 3rem;">Enabled</th>
          <th>Name</th>
          <th class="is-hidden-mobile">Setup</th>
          <th class="is-hidden-mobile">Up</th>
          <th class="is-hidden-mobile">Down</th>
          <th style="width: 6rem;">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="app in apps" :key="app.name">
          <td>
            <label class="toggle-switch">
              <input type="checkbox" :checked="app.enabled" @change="onToggleApp(app)" />
              <span class="toggle-slider round"></span>
            </label>
          </td>
          <td>
            <span class="icon-text">
              <span class="icon has-text-info"><i class="fas fa-cube"></i></span>
              <span class="has-text-weight-medium">{{ app.name }}</span>
            </span>
          </td>
          <td class="is-hidden-mobile cmd-count">
            <span class="tag is-light">{{ (app.setupCommands || []).length }} cmd{{ (app.setupCommands || []).length !== 1 ? 's' : '' }}</span>
          </td>
          <td class="is-hidden-mobile cmd-count">
            <span class="tag is-light is-success">{{ (app.upCommands || []).length }} cmd{{ (app.upCommands || []).length !== 1 ? 's' : '' }}</span>
          </td>
          <td class="is-hidden-mobile cmd-count">
            <span class="tag is-light is-warning">{{ (app.downCommands || []).length }} cmd{{ (app.downCommands || []).length !== 1 ? 's' : '' }}</span>
          </td>
          <td>
            <div class="buttons has-addons">
              <button type="button" class="button is-small is-info is-light"
                :id="'app-edit-' + app.name"
                @click="openEdit(app)" title="Edit app">
                <span class="icon"><i class="fas fa-pencil-alt"></i></span>
              </button>
              <button type="button" class="button is-small is-danger is-light"
                :id="'app-delete-' + app.name"
                @click="deleteTarget = app.name" title="Delete app">
                <span class="icon"><i class="fas fa-trash"></i></span>
              </button>
            </div>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Error banner -->
    <div v-if="errorMsg" class="notification is-danger is-light mt-3">
      <button class="delete" @click="errorMsg = ''"></button>
      {{ errorMsg }}
    </div>
  </div>
</template>

<script>
// Apps Manager Component
export default {
  name: 'apps-manager',
  props: {
    apps: {
      type: Array,
      default: () => [],
    },
  },
  emits: ['update:apps'],
  components: {
    'app-edit': Vue.defineAsyncComponent(() => ComponentLoader.import('app/app-edit')),
    'confirm-delete': Vue.defineAsyncComponent(() => ComponentLoader.import('core/confirm-delete')),
  },
  data() {
    return {
      showEdit: false,
      editingApp: null,
      deleteTarget: null,
      errorMsg: '',
    };
  },
  methods: {
    openAdd() {
      this.editingApp = null;
      this.showEdit = true;
    },
    openEdit(app) {
      this.editingApp = { ...app };
      this.showEdit = true;
    },
    closeEdit() {
      this.showEdit = false;
      this.editingApp = null;
    },
    async onSaveApp(appData) {
      this.showEdit = false;
      try {
        const resp = await fetch('/api/apps/save', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(appData),
        });
        if (!resp.ok) {
          const msg = await resp.text();
          this.errorMsg = `Save failed: ${msg}`;
        }
        // SSE will push updated list
      } catch (e) {
        this.errorMsg = `Save failed: ${e.message}`;
      }
    },
    async onDeleteApp() {
      const name = this.deleteTarget;
      this.deleteTarget = null;
      try {
        const resp = await fetch(`/api/apps/${name}/delete`, { method: 'POST' });
        if (!resp.ok) {
          const msg = await resp.text();
          this.errorMsg = `Delete failed: ${msg}`;
        }
        // SSE will push updated list
      } catch (e) {
        this.errorMsg = `Delete failed: ${e.message}`;
      }
    },
    async onToggleApp(app) {
      const cmd = app.enabled ? 'disable' : 'enable';
      try {
        const resp = await fetch(`/api/apps/${app.name}/${cmd}`, { method: 'POST' });
        if (!resp.ok) {
          const msg = await resp.text();
          this.errorMsg = `Toggle failed: ${msg}`;
        }
        // SSE will push updated list
      } catch (e) {
        this.errorMsg = `Toggle failed: ${e.message}`;
      }
    },
  },
}
</script>

<style>
.apps-manager {
  padding: 0.25rem;
}

.apps-table td {
  vertical-align: middle;
}

.cmd-count {
  text-align: center;
}
</style>
