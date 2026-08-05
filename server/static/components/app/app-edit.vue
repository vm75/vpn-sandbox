<template>
  <div v-if="isVisible" class="modal is-active" role="dialog" aria-modal="true" aria-labelledby="app-edit-title">
    <div class="modal-background" @click="cancel"></div>
    <div class="modal-card app-edit-card">
      <header class="modal-card-head">
        <p id="app-edit-title" class="modal-card-title">{{ isNew ? 'Add App' : 'Edit App' }}</p>
        <button type="button" class="delete" aria-label="close" @click="cancel"></button>
      </header>
      <section class="modal-card-body">
        <form @submit.prevent>
          <!-- Quick Add (Only show when creating new app) -->
          <div v-if="isNew" class="field is-horizontal">
            <div class="field-label is-normal">
              <label class="label">Quick Add</label>
            </div>
            <div class="field-body">
              <div class="field">
                <div class="control">
                  <div class="select is-fullwidth">
                    <select v-model="selectedPreset" @change="applyPreset">
                      <option value="">-- Select a preset (Optional) --</option>
                      <option v-for="preset in presets" :key="preset.name" :value="preset.name">
                        {{ preset.name }}
                      </option>
                    </select>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Name -->
          <div class="field is-horizontal">
            <div class="field-label is-normal">
              <label class="label">Name</label>
            </div>
            <div class="field-body">
              <div class="field">
                <div class="control">
                  <input id="app-edit-name" class="input" type="text" v-model="draft.name"
                    :readonly="!isNew" :class="{ 'is-static': !isNew }"
                    placeholder="e.g. deluge" />
                </div>
                <p v-if="nameError" class="help is-danger">{{ nameError }}</p>
              </div>
            </div>
          </div>

          <!-- Setup Commands -->
          <div class="field is-horizontal">
            <div class="field-label is-normal">
              <label class="label">Setup Commands</label>
            </div>
            <div class="field-body">
              <div class="field">
                <div class="control">
                  <textarea id="app-edit-setup" class="textarea cmd-textarea"
                    v-model="setupText"
                    placeholder="One command per line&#10;e.g. apt-get -y install deluge"
                    rows="3">
                  </textarea>
                </div>
                <p class="help">Runs when first saved or changed, and once per new container</p>
              </div>
            </div>
          </div>

          <!-- Up Commands -->
          <div class="field is-horizontal">
            <div class="field-label is-normal">
              <label class="label">Up Commands</label>
            </div>
            <div class="field-body">
              <div class="field">
                <div class="control">
                  <textarea id="app-edit-up" class="textarea cmd-textarea"
                    v-model="upText"
                    placeholder="One command per line&#10;e.g. deluged -c /data/config/deluge &"
                    rows="3">
                  </textarea>
                </div>
                <p class="help">Runs when the VPN tunnel comes up</p>
              </div>
            </div>
          </div>

          <!-- Down Commands -->
          <div class="field is-horizontal">
            <div class="field-label is-normal">
              <label class="label">Down Commands</label>
            </div>
            <div class="field-body">
              <div class="field">
                <div class="control">
                  <textarea id="app-edit-down" class="textarea cmd-textarea"
                    v-model="downText"
                    placeholder="One command per line&#10;e.g. pkill -9 deluged"
                    rows="3">
                  </textarea>
                </div>
                <p class="help">Runs when the VPN tunnel goes down</p>
              </div>
            </div>
          </div>
        </form>
      </section>
      <footer class="modal-card-foot">
        <button type="button" class="button" @click="cancel">Cancel</button>
        <button type="button" class="button is-success" @click="save">
          {{ isNew ? 'Add App' : 'Save Changes' }}
        </button>
      </footer>
    </div>
  </div>
</template>

<script>
// App Edit Modal Component
export default {
  name: 'app-edit',
  props: {
    show: {
      type: Boolean,
      default: false,
    },
    app: {
      type: Object,
      default: null,
    },
  },
  emits: ['cancel', 'save'],
  data() {
    return {
      isVisible: false,
      isNew: true,
      nameError: '',
      selectedPreset: '',
      presets: [
        {
          name: 'deluge',
          setupCommands: [
            'mkdir -p /data/config/deluge /downloads/completed /downloads/incomplete /downloads/watch',
            'apt-get update',
            'apt-get install -y deluge deluge-web'
          ],
          upCommands: [
            'deluged -c /data/config/deluge -l /data/var/deluge.log --logrotate 1024 &',
            'deluge-web -c /data/config/deluge -l /data/var/deluge-web.log --logrotate 1024 &'
          ],
          downCommands: [
            'pkill -9 deluged',
            'pkill -9 deluge-web'
          ]
        },
        {
          name: 'jackett',
          setupCommands: [
            'apt-get update && apt-get install -y wget',
            'mkdir -p /data/apps/jackett /data/config/Jackett',
            'wget -qO- https://github.com/Jackett/Jackett/releases/latest/download/Jackett.Binaries.LinuxAMDX64.tar.gz | tar -xz -C /data/apps/jackett --strip-components=1'
          ],
          upCommands: [
            'XDG_CONFIG_HOME=/data/config DisableRootWarning=true /data/apps/jackett/jackett &'
          ],
          downCommands: [
            'pkill -9 jackett'
          ]
        },
        {
          name: 'storm',
          setupCommands: [
            'apt-get update && apt-get install -y wget',
            'mkdir -p /data/apps/storm /data/config/storm',
            'wget -qO /data/apps/storm/storm https://github.com/relvacode/storm/releases/latest/download/storm-linux-amd64',
            'chmod +x /data/apps/storm/storm'
          ],
          upCommands: [
            'XDG_CONFIG_HOME=/data/config /data/apps/storm/storm -H 127.0.0.1 --deluge-version=v2 -u localclient -p YOUR_PASSWORD_HERE &'
          ],
          downCommands: [
            'pkill -f /data/apps/storm/storm'
          ]
        },
        {
          name: 'flaresolverr',
          setupCommands: [
            'apt-get update && apt-get install -y wget xvfb chromium',
            'mkdir -p /data/apps/flaresolverr /data/config/flaresolverr',
            'wget -qO- https://github.com/FlareSolverr/FlareSolverr/releases/latest/download/flaresolverr_linux_x64.tar.gz | tar -xz -C /data/apps/flaresolverr --strip-components=1'
          ],
          upCommands: [
            'LOG_LEVEL=info LOG_HTML=false HEADLESS=true XDG_CONFIG_HOME=/data/config HOME=/data/config/flaresolverr /data/apps/flaresolverr/flaresolverr &'
          ],
          downCommands: [
            'pkill -f flaresolverr'
          ]
        }
      ],
      draft: { name: '', enabled: true, setupCommands: [], upCommands: [], downCommands: [] },
      setupText: '',
      upText: '',
      downText: '',
    };
  },
  watch: {
    show(val) {
      if (val) this.open(this.app);
    },
    app(val) {
      if (this.isVisible) this.open(val);
    },
  },
  methods: {
    open(app) {
      this.isNew = !app || !app.name;
      this.nameError = '';
      this.selectedPreset = '';
      if (app) {
        this.draft = {
          name: app.name || '',
          enabled: app.enabled !== false,
          setupCommands: app.setupCommands || [],
          upCommands: app.upCommands || [],
          downCommands: app.downCommands || [],
        };
      } else {
        this.draft = { name: '', enabled: true, setupCommands: [], upCommands: [], downCommands: [] };
      }
      this.setupText = this.draft.setupCommands.join('\n');
      this.upText = this.draft.upCommands.join('\n');
      this.downText = this.draft.downCommands.join('\n');
      this.isVisible = true;
    },
    applyPreset() {
      if (!this.selectedPreset) return;
      const preset = this.presets.find(p => p.name === this.selectedPreset);
      if (preset) {
        this.draft.name = preset.name;
        this.setupText = preset.setupCommands.join('\n');
        this.upText = preset.upCommands.join('\n');
        this.downText = preset.downCommands.join('\n');
      }
    },
    cancel() {
      this.isVisible = false;
      this.$emit('cancel');
    },
    save() {
      const name = this.draft.name.trim();
      if (!name) {
        this.nameError = 'Name is required';
        return;
      }
      if (!/^[a-zA-Z0-9_-]+$/.test(name)) {
        this.nameError = 'Name may only contain letters, numbers, hyphens and underscores';
        return;
      }
      this.nameError = '';
      const toLines = (text) => text.split('\n').map(l => l.trim()).filter(l => l.length > 0);
      const result = {
        name,
        enabled: this.draft.enabled,
        setupCommands: toLines(this.setupText),
        upCommands: toLines(this.upText),
        downCommands: toLines(this.downText),
      };
      this.isVisible = false;
      this.$emit('save', result);
    },
  },
}
</script>

<style>
.app-edit-card {
  width: 680px;
  max-width: 95vw;
}

.cmd-textarea {
  font-family: monospace;
  font-size: 0.875rem;
  resize: vertical;
}
</style>
