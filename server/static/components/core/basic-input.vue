<template>
  <div>
    <input v-if="type === 'string'" :id="id" class="input" v-model="internalValue" :placeholder="placeholder"
      @input="emitInput">
    <textarea v-if="type === 'text'" :id="id" class="textarea" v-model="internalValue" :placeholder="placeholder"
      @input="emitInput">
    </textarea>
    <input v-if="type === 'int'" :id="id" class="input" v-model.number="internalValue" :placeholder="placeholder"
      type="number" step="1" @input="emitInput">
    <input v-if="type === 'float'" :id="id" class="input" v-model.number="internalValue" :placeholder="placeholder"
      type="number" step="any" @input="emitInput">
    <toggle v-if="type === 'checkbox'" :id="id" :type="'checkbox'" v-model:value="internalValue" @change="emitInput">
    </toggle>
    <toggle v-if="type === 'switch'" :id="id" :type="'switch'" v-model:value="internalValue" @change="emitInput">
    </toggle>
    <binary v-if="type === 'yes-no'" :id="id" :type="'yes-no'" v-model:value="internalValue" @change="emitInput">
    </binary>
    <binary v-if="type === 'on-off'" :id="id" :type="'on-off'" v-model:value="internalValue" @change="emitInput">
    </binary>
    <binary v-if="type === 'true-false'" :id="id" :type="'true-false'" v-model:value="internalValue"
      @change="emitInput">
    </binary>
    <input v-if="type === 'time'" :id="id" class="input" v-model="internalValue" :placeholder="placeholder"
      @blur="validateData" @input="emitInput">
  </div>
</template>

<script>
// Basic Input Component
export default {
  name: 'basic-input',
  props: {
    id: {
      type: String,
      required: true
    },
    type: {
      type: String,
      required: true
    },
    value: {
      type: [String, Number, Boolean],
      default: ''
    },
    placeholder: {
      type: String,
      default: ''
    },
    options: {
      type: Array,
      default: () => []
    },
  },
  components: {
    'binary': Vue.defineAsyncComponent(() => ComponentLoader.import('core/binary')),
    'toggle': Vue.defineAsyncComponent(() => ComponentLoader.import('core/toggle'))
  },
  data() {
    return {
      internalValue: this.value
    }
  },
  watch: {
    value(newValue) {
      this.internalValue = newValue; // Sync with parent when prop changes
    }
  },
  methods: {
    emitInput() {
      this.$emit('update:value', this.internalValue); // Emit value back to parent
    },
    validateData() {
      const regex = /^\d+[smhd]$/;
      if (!regex.test(this.internalValue)) {
        alert(`Invalid time format. Please use an integer followed by s, m, h, or d.`);
        this.internalValue = ''; // Clear the invalid input
      }
    },
    async init() {
    },
  },
  mounted() {
    this.init();
  }
}

</script>
