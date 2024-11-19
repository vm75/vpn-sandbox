<template>
  <div class="file-browser">
    <aside class="menu">
      <p class="menu-label">Files</p>
      <ul class="menu-list">
        <li v-for="item in items" :key="item.name">
          <a @click.prevent="selectItem(item)">
            <span v-if="item.isDir">📁</span>
            <span v-else>📄</span>
            {{ item.name }}
          </a>
        </li>
      </ul>
    </aside>
  </div>
</template>

<script>
// File Browser Component
export default {
  name: 'file-browser',
  props: {
    apiEndpoint: {
      type: String,
      required: true,
    },
  },
  data() {
    return {
      items: [],
    };
  },
  methods: {
    fetchItems(path = '') {
      fetch(`${this.apiEndpoint}?path=${path}`)
        .then((response) => response.json())
        .then((data) => {
          this.items = data;
        })
        .catch((error) => console.error('Error fetching items:', error));
    },
    selectItem(item) {
      if (item.isDir) {
        this.fetchItems(item.path);
      } else {
        this.$emit('file-selected', item.path);
      }
    },
  },
  mounted() {
    this.fetchItems();
  },
}
</script>

<style scoped>
.file-browser {
  padding: 10px;
  background: #f5f5f5;
  height: 100%;
  overflow-y: auto;
}

.menu-list a {
  cursor: pointer;
}

.menu-list a:hover {
  background-color: #f0f0f0;
}
</style>
