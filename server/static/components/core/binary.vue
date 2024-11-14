<template>
  <div class="select is-fullwidth">
    <select :id="id" v-model="internalValue" @change="emitInput">
      <option :value="true">{{ trueStr }}</option>
      <option :value="false">{{ falseStr }}</option>
    </select>
  </div>
</template>

<script>
export default {
  name: 'binary',
  props: ['id', 'type', 'value'], // type can be yes-no, on-off, true-false
  data() {
    return {
      internalValue: this.toBool(this.value), // Local copy of value for editing
      trueStr: 'true',
      falseStr: 'false',
    };
  },
  methods: {
    emitInput() {
      // Emit value back to parent
      this.$emit('update:value', this.fromBool(this.internalValue));
    },

    toBool(value) {
      // Convert the incoming value to a boolean
      if (['yes', 'on', 'true', 1, true].includes(value)) {
        return true;
      }
      return false;
    },

    fromBool(value) {
      // Convert the boolean back to the correct type-based string
      switch (this.type) {
        case 'yes-no':
          return value ? 'yes' : 'no';
        case 'on-off':
          return value ? 'on' : 'off';
        case 'true-false':
          return value ? 'true' : 'false';
        default:
          return value;
      }
    },

    async init() {
      this.internalValue = this.toBool(this.value);
      this.trueStr = this.fromBool(true);
      this.falseStr = this.fromBool(false);
      if (this.fromBool(this.internalValue) !== this.value) {
        this.$emit('update:value', this.fromBool(this.internalValue));
        this.$emit('change', this.fromBool(this.internalValue));
      }
    },
  },
  mounted() {
    this.init();
  },
  watch: {
    value(newValue) {
      // Whenever the value prop changes, reinitialize
      this.internalValue = this.toBool(newValue);
    }
  }
};

</script>
