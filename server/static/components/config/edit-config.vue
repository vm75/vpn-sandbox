<template>
  <div v-if="isVisible">
    <div class="modal is-active">
      <div class="modal-background"></div>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ config.name ? 'Edit Config' : 'New Config' }}</p>
          <button class="delete" aria-label="close" @click="cancel"></button>
        </header>
        <section class="modal-card-body">
          <div class="field">
            <legend class="label">Config Name</legend>
            <div class="control">
              <input class="input" v-model="config.name" placeholder="Config Name" />
            </div>
          </div>
          <div class="field">
            <legend class="label">Config Type</legend>
            <div class="control">
              <div class="select">
                <select v-model="config.templateName">
                  <option value="custom">Custom Config</option>
                  <option v-for="template in templates" :key="template.name" :value="template.name">
                    {{ template.name }}
                  </option>
                </select>
              </div>
            </div>
          </div>

          <!-- Template-based fields -->
          <div v-if="config.templateName !== 'custom'">
            <dynamic-form :config="dynamicConfig" v-model:data="config.fields" @update:data="onDataUpdate">
            </dynamic-form>
          </div>

          <!-- Custom data textarea -->
          <div v-if="config.templateName === 'custom'" class="field">
            <legend class="label">Custom Config Data</legend>
            <div class="control">
              <textarea class="textarea" v-model="config.data" placeholder="Custom Config Data"></textarea>
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
  name: "edit-config",
  props: {
    templates: {
      type: Array,
      required: true
    },
    name: {
      type: String,
      required: true
    },
    templateName: {
      type: String,
      required: true
    },
    fields: {
      type: Array,
      required: true
    },
    configData: {
      type: Object,
      default: () => ({}) // Default to an empty object if not provided
    },
    showOnLoad: {
      type: Boolean,
      default: false // Default to false if not provided
    }
  },
  components: {
    'dynamic-form': Vue.defineAsyncComponent(() => Component.import('components/core/dynamic-form')),
  },
  data() {
    return {
      isVisible: this.showOnLoad,
      config: {
        name: this.name,
        templateName: this.templateName,
        fields: this.fields,
        data: this.configData
      }
    }
  },
  methods: {
    // Used to show the modal using component ref
    show(name, templateName, fields, configData) {
      this.isVisible = true;
      this.config = { name: name, templateName: templateName, fields: fields, data: configData };
    },
    cancel() {
      this.isVisible = false;
      this.$emit('cancel');
    },
    save() {
      this.isVisible = false;
      this.$emit('save', this.config);
    },
    onDataUpdate(data) {
      this.config.fields = data;
    }
  },
  computed: {
    dynamicConfig() {
      if (this.config.templateName === "custom") {
        return [];
      }

      const template = this.templates.find(t => t.name === this.config.templateName);
      const content = template ? template.content : "";

      const regex = /\{\{(\w+)(:(\w+))?\}\}/g;
      const matches = content.matchAll(regex);
      const fields = [];
      for (const match of matches) {
        fields.push({ name: match[1], type: match[3] || "string" });
      }

      return fields;
    },
  }
}
</script>
