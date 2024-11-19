<template>
  <div class="file-viewer">
    <div v-if="content" class="card">
      <div class="card-content">
        <div class="content">
          <textarea class="textarea" readonly>{{ content }}</textarea>
        </div>
      </div>
    </div>
    <div v-else class="notification is-warning">
      No file selected
    </div>
  </div>
</template>

<script>
// File Viewer Component
export default {
  name: 'file-viewer',
  props: {
    apiEndpoint: {
      type: String,
      required: true,
    },
    filePath: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      content: null,
    };
  },
  watch: {
    filePath: {
      immediate: true,
      handler(newPath) {
        if (newPath) {
          this.fetchContent(newPath);
        } else {
          this.content = null;
        }
      },
    },
  },
  methods: {
    fetchContent(path) {
      fetch(`${this.apiEndpoint}?path=${path}`)
        .then((response) => response.text())
        .then((data) => {
          this.content = data;
        })
        .catch((error) => console.error('Error fetching file content:', error));
    },
  },
}
</script>

<style scoped>
.file-viewer {
  padding: 10px;
  background: #fff;
  height: 100%;
  overflow-y: auto;
}

pre.scrollable {
  white-space: pre;
  overflow-x: auto;
  /* Add horizontal scrollbar for long lines */
  word-wrap: normal;
  /* Prevent breaking words */
  max-height: 100%;
  /* Constrain to container */
  background-color: #f5f5f5;
  border: 1px solid #ddd;
  padding: 10px;
}
</style>