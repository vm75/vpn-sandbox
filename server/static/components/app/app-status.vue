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
          <icon v-if="running" :icon="mainIcon" banner="assets/locked.svg"></icon>
          <icon v-else-if="enabled" :icon="mainIcon" banner="assets/failed.svg"></icon>
          <icon v-else :icon="mainIcon" banner="assets/unlocked.svg"></icon>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
// App Status Component
export default {
  name: 'app-status',
  props: {
    name: {
      type: String,
      required: true
    },
    displayName: {
      type: String,
      required: true
    },
    enabled: {
      type: Boolean,
      required: true
    },
    running: {
      type: Boolean,
      required: true
    },
    failure: {
      type: String,
      default: ''
    }
  },
  emits: ['toggleModule'],
  data() {
    return {
      mainIcon: 'assets/' + this.name + '.svg',
      isEnabled: this.enabled,
    }
  },
  watch: {
    enabled(value) {
      this.isEnabled = value;
    }
  },
  components: {
    'icon': Vue.defineAsyncComponent(() => ComponentLoader.import('core/icon')),
    'basic': Vue.defineAsyncComponent(() => ComponentLoader.import('core/basic-input')),
  },
}
</script>

<style></style>
