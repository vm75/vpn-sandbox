<template>
  <div class="columns is-gapless file-explorer">
    <!-- Left Pane: File Explorer -->
    <div class="column is-3">
      <div class="box" style="height: 30em; overflow-y: auto;">
        <nav class="panel">
          <div class="panel-block">
            <ul>
              <li v-for="(item, index) in fileTree" :key="index" @click="fetchFileContent(item.path)"
                class="has-text-weight-semibold">
                <a>
                  <span v-if="item.isDir">📁</span>
                  <span v-else>📄</span>
                  {{ item.name }}
                </a>
              </li>
            </ul>
          </div>
        </nav>
      </div>
    </div>

    <!-- Right Pane: File Viewer -->
    <div class="column is-9">
      <div class="box" style="height: 30em; overflow-y: auto;">
        <div v-if="loading" class="has-text-centered">
          <p>Loading...</p>
        </div>
        <div v-else>
          <pre v-if="fileContent" class="content">{{ fileContent }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
// File Explorer Component
export default {
  name: 'file-explorer',
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
  data() {
    return {
      fileTree: [],
      fileContent: null,
      loading: false,
    };
  },
  methods: {
    // Fetch the file tree from the server
    async fetchFileTree(path = '') {
      try {
        const response = await fetch(`${this.filesEndpoint}?path=${path}`);
        if (!response.ok) {
          throw new Error('Network response was not ok');
        }
        this.fileTree = await response.json();
      } catch (error) {
        console.error('Error fetching file tree:', error);
      }
    },
    // Fetch the content of the selected file
    async fetchFileContent(path) {
      this.loading = true;
      try {
        const response = await fetch(`${this.fileEndpoint}?path=${path}`);
        if (!response.ok) {
          throw new Error('Network response was not ok');
        }
        this.fileContent = await response.text();
      } catch (error) {
        console.error('Error fetching file content:', error);
      } finally {
        this.loading = false;
      }
    },
  },
  mounted() {
    // Initially fetch the root directory contents
    this.fetchFileTree();
  },
};
</script>

<style scoped>
.file-explorer .content {
  white-space: pre-wrap;
  word-wrap: break-word;
}
</style>
