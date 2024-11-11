const template = `
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
`;

export default {
  props: ["show", "title"],
  data() {
    return {
      text: 'This is the default text.', // Initial text
      isModalOpen: this.show,
    }
  },
  template: template,
  computed: {
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
