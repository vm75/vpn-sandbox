<template>
  <div class="file-browser-viewer columns is-gapless">
    <div class="column is-one-quarter">
      <file-browser :apiEndpoint="filesEndpoint" @file-selected="handleFileSelected" />
    </div>
    <div class="column">
      <file-viewer :apiEndpoint="fileEndpoint" :file-path="selectedFilePath" />
    </div>
  </div>
</template>

<script>
// File Browser Viewer Component
export default {
  name: 'file-browser-viewer',
  props: {
    filesEndpoint: {
      type: String,
      required: true,
    },
    fileEndpoint: {
      type: String,
      required: true,
    },
  },
  components: {
    'file-browser': Vue.defineAsyncComponent(() => ComponentLoader.import('core/file-browser')),
    'file-viewer': Vue.defineAsyncComponent(() => ComponentLoader.import('core/file-viewer')),
  },
  data() {
    return {
      selectedFilePath: '',
    };
  },
  methods: {
    handleFileSelected(path) {
      this.selectedFilePath = path;
    },
  },
}
</script>

<style scoped>
.file-browser-viewer {
  height: 100vh;
}
</style>
