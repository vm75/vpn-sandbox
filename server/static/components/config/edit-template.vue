<template>
  <div v-if="isVisible">
    <div class="modal is-active">
      <div class="modal-background"></div>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ template.name ? 'Edit Template' : 'New Template' }}</p>
          <button class="delete" aria-label="close" @click="cancel"></button>
        </header>
        <section class="modal-card-body">
          <div class="field">
            <legend class="label">Template Name</legend>
            <div class="control">
              <input class="input" v-model="template.name" placeholder="Template Name" />
            </div>
          </div>
          <div class="field">
            <legend class="label">Template Content</legend>
            <div class="control">
              <textarea class="textarea" v-model="template.content" placeholder="Template Content"></textarea>
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
  name: "edit-template",
  props: {
    name: {
      type: String,
      required: true
    },
    contents: {
      type: String,
      default: ''
    },
    showOnLoad: {
      type: Boolean,
      default: false // Default to false if not provided
    }
  },
  data() {
    return {
      isVisible: this.showOnLoad,
      template: {
        name: this.name,
        content: this.content,
      }
    }
  },
  methods: {
    show(name, content) {
      this.isVisible = true;
      this.template.name = name;
      this.template.content = content;
    },
    cancel() {
      this.isVisible = false;
      this.$emit('cancel');
    },
    save() {
      this.isVisible = false;
      this.$emit('save', this.template);
    },
  },
}
</script>
