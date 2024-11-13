<template>
  <div v-if="isVisible">
    <div class="modal is-active">
      <div class="modal-background"></div>
      <div class="modal-card">
        <header class="modal-card-head">
          <p class="modal-card-title">{{ title }}</p>
          <button class="delete" aria-label="close" @click="cancel"></button>
        </header>
        <section class="modal-card-body">
          <form @submit.prevent>
            <div class="field is-horizontal">
              <div class="field-label is-normal">
                <label class="label">Value</label>
              </div>
              <div class="field-body">
                <div class="field">
                  <basic-input :type="type" v-model:value="newMapValue" :placeholder="placeholder">
                  </basic-input>
                </div>
              </div>
            </div>
          </form>
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
  name: "edit-item",
  props: ["title", "type", "placeholder", "showOnLoad", "initialValue"],
  components: {
    'basic-input': Vue.defineAsyncComponent(() => importComponent('components/core/basic-input'))
  },
  data() {
    return {
      isVisible: this.showOnLoad || false,
      newMapValue: this.initialValue || '',
    }
  },
  methods: {
    show(value) {
      this.newMapValue = value;
      this.isVisible = true;
    },
    cancel() {
      this.isVisible = false;
      this.$emit('cancel');
    },
    save() {
      this.isVisible = false;
      this.$emit('save', this.newMapValue);
    },
  },
}
</script>
