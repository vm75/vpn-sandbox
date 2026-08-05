<template>
  <div class="columns file-explorer">
    <!-- Left Pane: File Explorer -->
    <div class="column is-3 explorer-sidebar">
      <div class="panel is-info" style="height: 30em; display: flex; flex-direction: column;">
        <p class="panel-heading">
          <span class="icon-text">
            <span class="icon"><i class="fas fa-folder-open"></i></span>
            <span>Files</span>
          </span>
        </p>
        <div class="panel-list" style="flex-grow: 1; overflow-y: auto; padding: 0.5rem 0;">
          <ul class="file-tree-list">
            <li v-for="(item, index) in fileTree" :key="index"
               @click="fetchFileContent(item)"
               class="file-tree-item is-flex is-align-items-center"
               :class="{ 'is-active': selectedFile && selectedFile.path === item.path }">
              <span class="file-icon mr-2">
                <i class="fas" :class="item.isDir ? 'fa-folder has-text-warning' : 'fa-file-alt has-text-info'"></i>
              </span>
              <span class="file-tree-name" style="white-space: nowrap; overflow: hidden; text-overflow: ellipsis;" :title="item.name">{{ item.name }}</span>
            </li>
            <li v-if="fileTree.length === 0" class="has-text-grey-light is-flex is-justify-content-center py-5">
              <div class="has-text-centered">
                <i class="fas fa-folder-open fa-2x mb-2"></i>
                <p>No files found</p>
              </div>
            </li>
          </ul>
        </div>
      </div>
    </div>

    <!-- Right Pane: File Viewer -->
    <div class="column is-9 file-viewer-container">
      <div class="box p-0 file-viewer-box is-flex is-flex-direction-column" style="height: 30em;">
        <!-- Header -->
        <div class="file-header p-3 has-background-white-ter is-flex is-justify-content-space-between is-align-items-center" style="border-bottom: 1px solid #dbdbdb; border-radius: 6px 6px 0 0;">
          <div class="is-flex is-align-items-center">
            <span class="icon is-small mr-2 has-text-grey">
              <i class="fas fa-file-code"></i>
            </span>
            <span class="has-text-weight-semibold has-text-grey-dark">
              {{ selectedFile ? selectedFile.name : 'Select a file' }}
            </span>
          </div>
          <div class="is-flex align-items-center">
             <button v-if="selectedFile" class="button is-small is-light" @click="fetchFileContent(selectedFile)" :class="{'is-loading': loading}">
              <span class="icon is-small"><i class="fas fa-sync-alt"></i></span>
            </button>
          </div>
        </div>

        <!-- Content -->
        <div class="file-content-wrapper p-3" style="flex-grow: 1; overflow-y: auto; background-color: #fafafa; border-radius: 0 0 6px 6px;" ref="fileViewer">
          <div v-if="loading && !fileContent" class="is-flex is-justify-content-center is-align-items-center" style="height: 100%">
            <span class="icon is-large has-text-info">
              <i class="fas fa-circle-notch fa-spin fa-2x"></i>
            </span>
          </div>
          <div v-else-if="!selectedFile" class="is-flex is-justify-content-center is-align-items-center has-text-grey-light" style="height: 100%">
            <div class="has-text-centered">
               <i class="fas fa-file-signature fa-3x mb-3"></i>
               <p>Select a file to view its contents</p>
            </div>
          </div>
          <pre v-else class="file-content">{{ fileContent || 'File is empty' }}</pre>
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
      selectedFile: null,
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
        const data = await response.json();
        this.fileTree = data.sort((a, b) => a.name.localeCompare(b.name));
      } catch (error) {
        console.error('Error fetching file tree:', error);
      }
    },
    // Fetch the content of the selected file
    async fetchFileContent(item) {
      if (item.isDir) {
        return; // Directory traversal can be implemented here if needed
      }

      this.selectedFile = item;
      this.loading = true;
      try {
        const response = await fetch(`${this.fileEndpoint}?path=${item.path}`);
        if (!response.ok) {
          throw new Error('Network response was not ok');
        }
        this.fileContent = await response.text();
      } catch (error) {
        console.error('Error fetching file content:', error);
        this.fileContent = 'Error fetching file content.';
      } finally {
        this.loading = false;
        this.$nextTick(() => {
          if (this.$refs.fileViewer) {
            this.$refs.fileViewer.scrollTop = this.$refs.fileViewer.scrollHeight;
          }
        });
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
.file-explorer {
  margin-top: 0.5rem;
}
.file-explorer .file-content {
  white-space: pre-wrap;
  word-wrap: break-word;
  background-color: transparent;
  padding: 0;
  font-family: 'Consolas', 'Courier New', monospace;
  font-size: 0.9em;
  color: #333;
}
.file-viewer-box {
  box-shadow: 0 0.5em 1em -0.125em rgba(10, 10, 10, 0.1), 0 0px 0 1px rgba(10, 10, 10, 0.02);
}
.file-tree-list {
  list-style: none;
  margin: 0;
  padding: 0;
}
.file-tree-item {
  padding: 0.15rem 0.75rem;
  cursor: pointer;
  transition: background-color 0.2s ease;
  font-size: 0.9em;
  line-height: 1.25;
  border: 0;
  border-radius: 0;
}
.file-tree-item.is-active {
  color: #1677b8;
  font-weight: 600;
}
.file-tree-item:hover {
  background-color: #f5f5f5;
}
</style>
