<template>
  <confirm-delete :show="removeIndex !== null" :itemName="removeItemName" @cancel="cancelRemoveItem"
    @confirm="confirmRemoveItem">
  </confirm-delete>
  <div v-if="listLocal.length > 0" class="b-table mb-4">
    <table class="table is-fullwidth is-hoverable align-middle" style="background-color: transparent;">
      <thead>
        <tr>
          <th class="has-text-grey-light is-uppercase is-size-7" style="border-bottom: 2px solid #f0f0f0;">{{ name }}</th>
          <th class="has-text-grey-light is-uppercase is-size-7 has-text-right" style="border-bottom: 2px solid #f0f0f0; width: 120px;">Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="(item, arrIndex) in listLocal" :key="arrIndex" style="transition: background-color 0.2s;">
          <td class="has-text-weight-medium is-vcentered">
            <div class="is-flex is-align-items-center">
              <span class="icon has-text-primary mr-3"><i class="fas fa-hdd"></i></span>
              {{ getDisplayString(arrIndex) }}
            </div>
          </td>
          <td class="has-text-right is-vcentered">
            <button type="button" class="button is-small is-light is-info is-rounded mr-2" @click="editItemLocal(arrIndex)" title="Edit">
              <span class="icon"><i class="fas fa-edit"></i></span>
            </button>
            <button type="button" class="button is-small is-light is-danger is-rounded" @click="removeItemLocal(arrIndex)" title="Delete">
              <span class="icon"><i class="fas fa-trash-alt"></i></span>
            </button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
  <div v-else class="notification is-light has-text-centered my-5" style="border-radius: 8px;">
    <div class="icon has-text-grey-light is-large mb-3"><i class="fas fa-folder-open fa-2x"></i></div>
    <p class="has-text-grey">No {{ name.toLowerCase() }}s found.</p>
  </div>

  <div v-if="addItem" class="has-text-right mt-2">
    <button type="button" class="button is-primary is-rounded px-4" @click="addItemLocal()">
      <span class="icon"><i class="fas fa-plus"></i></span>
      <span>Add {{ name.split(' ')[0] }}</span>
    </button>
  </div>
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
