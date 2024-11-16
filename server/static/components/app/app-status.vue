<template>
  <div class="field is-horizontal">
    <div class="field-label is-normal">
      <legend class="label">{{ displayName }}</legend>
    </div>
    <div class="field-body">
      <div class="field is-flex is-justify-content-space-around is-align-items-center">
        <!-- Switch -->
        <div class="control">
          <basic :id="name + '-enabled'" type="switch" v-model:value="isEnabled"
            @update:value="$emit('toggleModule', name)">
          </basic>
        </div>
        <!-- Icon with Status Banner -->
        <div class="control">
          <icon v-if="isRunning" :icon="mainIcon" banner="assets/locked.svg"></icon>
          <icon v-else :icon="mainIcon" banner="assets/unlocked.svg"></icon>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  props: {
    name: {
      type: String,
      required: true
    },
    displayName: {
      type: String,
      required: true
    },
    config: {
      type: Object,
      required: true
    }
  },
  emits: ['toggleModule'],
  data() {
    return {
      isEnabled: this.config.enabled,
      isRunning: this.config.running,
      mainIcon: 'assets/' + this.name + '.svg',
    }
  },
  components: {
    'icon': Vue.defineAsyncComponent(() => ComponentLoader.import('core/icon')),
    'basic': Vue.defineAsyncComponent(() => ComponentLoader.import('core/basic-input')),
  },
  watch: {
    config(newConfig) {
      this.isEnabled = newConfig.enabled;
      this.isRunning = newConfig.running;
    }
  }
}
</script>

<style></style>
