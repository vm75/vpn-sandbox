<template>
  <div v-if="isModalOpen">
    <div class="modal is-active">
      <div class="modal-background"></div>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ title }}</p>
        </header>
        <section class="modal-card-body">
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
// Modal Template
export default {
  name: "modal-template",
  props: {
    show: {
      type: Boolean,
      required: true
    },
    title: {
      type: String,
      required: true
    },
  },
  props: ["show", "title"],
  data() {
    return {
      text: 'This is the default text.', // Initial text
      isModalOpen: this.show,
    }
  },
  methods: {
    cancel() {
      this.$emit('cancel');
      this.isModalOpen = false;
    },
    save() {
      this.$emit('save', this.localText);
      this.isModalOpen = false;
    },
    async init() {
    }
  },
  mounted() {
    this.init();
  }
}
</script>
