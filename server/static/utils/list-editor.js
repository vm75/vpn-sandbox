const template = `
  <table v-if="listLocal.length > 0" class="table is-striped is-fullwidth">
    <thead>
      <tr>
        <th>Actions</th>
        <th>{{name}}</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="(item, arrIndex) in listLocal" :key="arrIndex">
        <td>
          <button class="button is-rounded is-small is-info is-light" @click="editItemLocal(arrIndex)">
          ✎
          </button>
          <button class="button is-rounded is-small is-danger is-light" @click="removeItemLocal(arrIndex)">
          🗑
          </button>
        </td>
        <td>
          {{getDisplayString(arrIndex)}}
        </td>
      </tr>
    </tbody>
  </table>
  <button class="button is-small is-info" @click="addItemLocal()">➕</button>
`;

export default {
  name: "edit-list",
  props: ["name", "list", "editItem", "addItem", "removeItem", "displayString"],
  data() {
    return {
      listLocal: this.list || [],
    }
  },
  template: template,
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
        const result = this.removeItem(index);
        if (result) {
          this.listLocal.splice(index, 1);
          this.emitData();
        }
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
}
