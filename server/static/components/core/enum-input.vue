<template>
  <div class="select">
    <select v-model="internalValue" @change="emitInput">
      <option v-for="option in options" :key="getKey(option)" :value="getKey(option)"
        :selected="getKey(option) === value">{{ getDisplayStr(option) }}
      </option>
    </select>
  </div>
</template>

<script>
export default {
  name: 'enum-input',
  props: ['value', 'options'],
  data() {
    return {
      internalValue: this.value, // Local copy of value for editing
    };
  },
  methods: {
    emitInput() {
      // Emit value back to parent
      this.$emit('update:value', this.internalValue);
    },
    getKey(option) {
      // if array, pick first element
      if (Array.isArray(option)) {
        return option[0];
      }
      return option;
    },
    getDisplayStr(option) {
      // if array, pick second element
      if (Array.isArray(option)) {
        return option[1];
      }
      return option;
    },
  }
}
</script>
