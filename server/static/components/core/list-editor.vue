<template>
  <confirm-delete :show="removeIndex !== null" :itemName="removeItemName" @cancel="cancelRemoveItem"
    @confirm="confirmRemoveItem">
  </confirm-delete>
  <table v-if="listLocal.length > 0" class="table is-striped is-fullwidth">
    <thead>
      <tr>
        <th>Actions</th>
        <th>{{ name }}</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="(item, arrIndex) in listLocal" :key="arrIndex">
        <td>
          <button type="button" class="button is-rounded is-small is-info is-light" @click="editItemLocal(arrIndex)">
            ✎
          </button>
          <button type="button" class="button is-rounded is-small is-danger is-light" @click="removeItemLocal(arrIndex)">
            🗑
          </button>
        </td>
        <td>
          {{ getDisplayString(arrIndex) }}
        </td>
      </tr>
    </tbody>
  </table>
  <button type="button" class="button is-small is-info" @click="addItemLocal()">➕</button>
</template>

<script>
// List Editor Component
export default {
  name: "list-editor",
  props: {
    name: {
      type: String,
      required: true
    },
    list: {
      type: Array,
      required: true
    },
    editItem: {
      type: Function,
      default: null
    },
    addItem: {
      type: Function,
      default: null
    },
    removeItem: {
      type: Function,
      default: null
    },
    displayString: {
      type: Function,
      default: null
    }
  }
  ,
  data() {
    return {
      listLocal: this.list || [],
      removeIndex: null,
    }
  },
  components: {
    'confirm-delete': Vue.defineAsyncComponent(() => ComponentLoader.import('core/confirm-delete')),
  },
  watch: {
    list(newList) {
      this.listLocal = newList || [];
    }
  },
  methods: {
    editItemLocal(index) {
      if (this.editItem) {
        const item = this.editItem(index);
        if (item) {
          this.listLocal[index] = item;
          this.emitData();
        }
      }
    },
    addItemLocal() {
      if (this.addItem) {
        const item = this.addItem();
        if (item) {
          this.listLocal.push(item);
          this.emitData();
        }
      }
    },
    removeItemLocal(index) {
      if (this.removeItem) {
        this.removeIndex = index;
      }
    },
    cancelRemoveItem() {
      this.removeIndex = null;
    },
    confirmRemoveItem() {
      const index = this.removeIndex;
      this.removeIndex = null;
      if (index === null || !this.removeItem) {
        return;
      }

      const result = this.removeItem(index);
      if (result) {
        this.listLocal.splice(index, 1);
        this.emitData();
      }
    },
    emitData() {
      this.$emit('update:list', this.listLocal);
    },
    getDisplayString(index) {
      if (this.displayString) {
        return this.displayString(this.listLocal[index]);
      }
      return this.listLocal[index];
    }
  },
  computed: {
    removeItemName() {
      if (this.removeIndex === null) {
        return '';
      }
      return this.getDisplayString(this.removeIndex);
    },
  },
}
</script>
