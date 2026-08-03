<template>
  <div v-if="show" class="modal is-active" role="dialog" aria-modal="true" aria-labelledby="confirm-delete-title">
    <div class="modal-background" @click="cancel"></div>
    <div class="modal-card">
      <header class="modal-card-head">
        <p id="confirm-delete-title" class="modal-card-title">Confirm deletion</p>
        <button type="button" class="delete" aria-label="close" @click="cancel"></button>
      </header>
      <section class="modal-card-body">
        Delete <strong>{{ itemName }}</strong>? This cannot be undone.
      </section>
      <footer class="modal-card-foot">
        <button type="button" class="button mx-auto" @click="cancel">Cancel</button>
        <button type="button" class="button is-danger mx-auto" @click="confirm">Delete</button>
      </footer>
    </div>
  </div>
</template>

<script>
export default {
  name: "confirm-delete",
  props: {
    show: {
      type: Boolean,
      required: true,
    },
    itemName: {
      type: String,
      required: true,
    },
  },
  methods: {
    cancel() {
      this.$emit('cancel');
    },
    confirm() {
      this.$emit('confirm');
    },
  },
}
</script>
