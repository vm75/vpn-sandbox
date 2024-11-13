<template>
  <edit-item ref="arrayItemModal" :title="editValueTitle" :type="editValueType" @save="onSaveValue">
  </edit-item>
  <edit-key-val ref="mapItemModal" :title="editKeyValueTitle" :type="editKeyValueType" @save="onSaveKeyValue">
  </edit-key-val>
  <div>
    <form @submit.prevent>
      <div v-for="(field, index) in fields" :key="index" class="mb-4">
        <div class="field is-horizontal">
          <div class="field-label">
            <label class="label">{{ field.name }}</label>
          </div>
          <div class="field-body">
            <div class="field is-small">
              <div class="control is-expanded">
                <!-- Basic Input Handling -->
                <basic-input v-if="isBasicType(field.type)" :type="field.type" v-model:value="formData[field.name]"
                  @change="onChange(field, formData[field.name])">
                  >
                </basic-input>

                <!-- Enum Input Handling -->
                <enum-input v-if="field.type === 'enum'" v-model:value="formData[field.name]" :options="field.options"
                  @change="onChange(field, formData[field.name])">
                  >
                </enum-input>

                <!-- Dynamic Array Handling -->
                <div v-if="isArray(field.type)" class="box">
                  <table v-if="formData[field.name].length > 0" class="table is-striped">
                    <thead>
                      <tr>
                        <th>Actions</th>
                        <th>Value</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="(item, arrIndex) in formData[field.name]" :key="arrIndex">
                        <td>
                          <button class="button is-rounded is-small is-info is-light"
                            @click="editArrayItem(field, arrIndex)">
                            ✎
                          </button>
                          <button class="button is-rounded is-small is-danger is-light"
                            @click="removeArrayItem(field, arrIndex)">
                            🗑
                          </button>
                        </td>
                        <td>
                          {{ formData[field.name][arrIndex] }}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                  <button class="button is-small is-info" @click="addArrayItem(field)">➕</button>
                </div>

                <!-- Dynamic Map Handling -->
                <div v-if="isMap(field.type)" class="box">
                  <table v-if="Object.keys(formData[field.name]).length > 0" class="table is-striped">
                    <thead>
                      <tr>
                        <th>Actions</th>
                        <th>Key</th>
                        <th>Value</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="(value, key) in formData[field.name]" :key="key">
                        <td>
                          <button class="button is-rounded is-small is-info is-light" @click="editMapItem(field, key)">✎
                          </button>
                          <button class="button is-rounded is-small is-danger is-light"
                            @click="removeMapItem(field, key)">🗑
                          </button>
                        </td>
                        <td>
                          {{ key }}
                        </td>
                        <td>
                          {{ formData[field.name][key] }}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                  <button class="button is-small is-info" @click="addMapItem(field)">➕</button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </form>
  </div>
</template>

<script>
export default {
  name: "dynamic-form",
  props: ["config", "data"],
  components: {
    'basic-input': Vue.defineAsyncComponent(() => importComponent('components/core/basic-input')),
    'enum-input': Vue.defineAsyncComponent(() => importComponent('components/core/enum-input')),
    'edit-key-val': Vue.defineAsyncComponent(() => importComponent('components/core/edit-key-val')),
    'edit-item': Vue.defineAsyncComponent(() => importComponent('components/core/edit-item')),
  },
  data() {
    return {
      // edit value
      currentArrayField: '',
      showEditValue: false,
      editValueTitle: '',
      editValueType: '',
      editValueIndex: -1,

      // edit key-value
      currentMapField: '',
      showEditKeyValue: false,
      editKeyValueTitle: '',
      editKeyValueType: '',
      editKeyValueKey: '',

      fields: [],
      formData: {},
    }
  },
  watch: {
    config(newValue) {
      this.init(newValue);
    },
  },
  methods: {
    isArray(type) {
      return type.startsWith('array_');
    },
    isMap(type) {
      return type.startsWith('map_');
    },
    isBasicType(type) {
      if (this.isArray(type) || this.isMap(type)) {
        return false;
      }
      return true;
    },
    getSubtype(type) {
      return type.split('_')[1];
    },
    addArrayItem(field) {
      this.currentArrayField = { name: field.name, type: field.type };
      this.editValueTitle = 'Add to ' + field.name;
      this.editValueType = this.getSubtype(field.type);
      this.editValueIndex = -1;

      this.$refs.arrayItemModal.show('');
    },
    editArrayItem(field, index) {
      this.currentArrayField = { name: field.name, type: field.type };
      this.editValueTitle = 'Edit ' + field.name;
      this.editValueType = this.getSubtype(field.type);
      this.editValueIndex = index;

      this.$refs.arrayItemModal.show(this.formData[field.name][index]);
    },
    removeArrayItem(field, index) {
      this.formData[field.name].splice(index, 1);
      this.emitData();
    },
    onSaveValue(value) {
      if (this.editValueIndex >= 0) {
        this.formData[this.currentArrayField.name][this.editValueIndex] = value;
      } else {
        this.formData[this.currentArrayField.name].push(value);
      }
      this.emitData();
    },
    addMapItem(field) {
      this.currentMapField = { name: field.name, type: field.type };
      this.editKeyValueTitle = 'Add to ' + field.name;
      this.editKeyValueType = this.getSubtype(field.type);
      this.editKeyValueKey = '';

      this.$refs.mapItemModal.show('', '');
    },
    editMapItem(field, key) {
      this.currentMapField = { name: field.name, type: field.type };
      this.editKeyValueTitle = 'Edit ' + field.name;
      this.editKeyValueType = this.getSubtype(field.type);
      this.editKeyValueKey = key;

      this.$refs.mapItemModal.show(key, this.formData[field.name][key]);
    },
    removeMapItem(field, key) {
      delete this.formData[field.name][key];
      this.emitData();
    },
    onSaveKeyValue(data) {
      if (this.editKeyValueKey) {
        this.formData[this.currentMapField.name][this.editValueIndex] = data.value;
      } else {
        this.formData[this.currentMapField.name][data.key] = data.value;
      }
      this.emitData();
    },
    onChange(field, value) {
      // this.formData[field.name] = value;
      this.emitData();
    },
    emitData() {
      this.$emit('update:data', { ...this.formData });
    },
    getInitialValue(type) {
      switch (type) {
        case 'string':
        case 'text':
          return '';
        case 'int':
          return 0;
        case 'float':
          return 0.0;
        case 'checkbox':
        case 'switch':
        case 'yes-no':
        case 'on-off':
        case 'true-false':
          return false;
        case 'time':
          return '';
      }
      return '';
    },
    init(config) {
      for (let field of config) {
        if (this.data.hasOwnProperty(field.name)) {
          this.formData[field.name] = this.data[field.name];
          continue;
        }
        if (this.isArray(field.type)) {
          this.formData[field.name] = [];
        } else if (this.isMap(field.type)) {
          this.formData[field.name] = {};
        } else {
          this.formData[field.name] = this.getInitialValue(field.type);
        }
      }
      this.fields = config;
    },
  },
  mounted() {
    this.init(this.config);
  }
}
</script>

<style></style>
