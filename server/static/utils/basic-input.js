const toggle = {
  template: `
    <div>
      <input v-if="type === 'checkbox'" type="checkbox" v-model="internalValue" @change="emitInput"/>
      <label v-if="type === 'switch'" class="switch">
        <input type="checkbox" v-model="internalValue" @change="emitInput"/>
        <span class="slider round"></span>
      </label>
    </div>
  `,
  name: 'toggle',
  props: ['type', 'value'],
  data() {
    return {
      internalValue: this.value, // Local copy of value for editing
    };
  },
  methods: {
    emitInput() {
      // Emit value back to parent
      this.$emit('update:value', this.internalValue);
    }
  },
  watch: {
    value(newValue) {
      this.internalValue = newValue; // Sync with parent when prop changes
    }
  },
  mounted() {
    const switchCssUrl = './utils/switch.css';

    if (this.type === 'switch' && !isLoaded(switchCssUrl)) {
      injectStyleUrl(switchCssUrl);
    }
  }
}

const binary = {
  template: `
    <div class="select is-fullwidth">
      <select v-model="internalValue" @change="emitInput">
        <option :value="true">{{ trueStr }}</option>
        <option :value="false">{{ falseStr }}</option>
      </select>
    </div>
  `,
  name: 'binary',
  props: ['type', 'value'], // type can be yes-no, on-off, true-false
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

const template = `
  <div>
    <input v-if="type === 'string'" class="input"
      v-model="internalValue"
      :placeholder="placeholder"
      @input="emitInput">
    <textarea
      v-if="type === 'text'"
      class="textarea"
      v-model="internalValue"
      :placeholder="placeholder"
      @input="emitInput">
    </textarea>
    <input v-if="type === 'int'" class="input"
      v-model.number="internalValue"
      :placeholder="placeholder"
      type="number"
      step="1"
      @input="emitInput">
    <input v-if="type === 'float'" class="input"
      v-model.number="internalValue"
      :placeholder="placeholder"
      type="number"
      step="any"
      @input="emitInput">
    <toggle v-if="type === 'checkbox'"
      :type="'checkbox'"
      v-model:value="internalValue"
      @change="emitInput">
    </toggle>
    <toggle v-if="type === 'switch'"
      :type="'switch'"
      v-model:value="internalValue"
      @change="emitInput">
    </toggle>
    <binary v-if="type === 'yes-no'"
      :type="'yes-no'"
      v-model:value="internalValue"
      @change="emitInput">
    </binary>
    <binary v-if="type === 'on-off'"
      :type="'on-off'"
      v-model:value="internalValue"
      @change="emitInput">
    </binary>
    <binary v-if="type === 'true-false'"
      :type="'true-false'"
      v-model:value="internalValue"
      @change="emitInput">
    </binary>
    <input v-if="type === 'time'" class="input"
      v-model="internalValue"
      :placeholder="placeholder"
      @blur="validateData"
      @input="emitInput">
  </div>
`;

export default {
  name: "basic-input",
  props: ["type", "value", "placeholder", "options"],
  components: {
    'toggle': toggle,
    'binary': binary,
    'toggle': toggle
  },
  data() {
    return {
      internalValue: this.value
    }
  },
  template: template,
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
