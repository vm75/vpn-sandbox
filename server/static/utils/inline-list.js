const template = `
  <div class="entries-input">
    <div v-for="(entry, index) in entries" :key="index" class="tag is-link">
      {{ entry }}
      <button
        class="delete is-small ml-1"
        @click.prevent="removeEntry(index)"
      ></button>
    </div>
    <input
      class="input"
      type="text"
      :placeholder="placeholder"
      v-model="input"
      @keydown.enter.prevent="addEntry"
      @keyup.space="addEntry"
    />
  </div>
`;

export default {
  template: template,
  props: {
    entries: {
      type: Array,
      default: () => [],
    },
    placeholder: {
      type: String,
      default: () => "Add an entry",
    },
    pattern: {
      type: RegExp,
      default: () => /.+/, // Default to accept any non-empty tag
    },
    type: {
      type: String,
      default: null,
    },
  },
  data() {
    return {
      input: "",
    }
  },
  computed: {
  },
  methods: {
    addEntry() {
      const newEntries = this.input.trim().split(/\s+/);
      const validEntries = newEntries.filter((entry) => this.validator.test(entry) && !this.entries.includes(entry));
      this.$emit("update:entries", [...this.entries, ...validEntries]); // Emit updated entries array
      this.input = "";
    },
    removeEntry(index) {
      const updatedEntries = [...this.entries];
      updatedEntries.splice(index, 1);
      this.$emit("update:entries", updatedEntries); // Emit updated entries array
    },
  },
  computed: {
    validator: function () {
      switch (this.type) {
        case "subnet":
          return new RegExp(/^(([0-9]{1,3}\.){3}[0-9]{1,3})\/([0-9]{1,2})$/);
        case "ipv4":
          return new RegExp(/^(([0-9]{1,3}\.){3}[0-9]{1,3})$/);
        case "int":
          return new RegExp(/^-?\d+$/);
        case "float":
          return new RegExp(/^-?\d+(\.\d+)?$/);
        case "email":
          return new RegExp(/^[^\s@]+@[^\s@]+\.[^\s@]+$/);
        case "url":
          return new RegExp(/^(https?:\/\/)?([\da-z\.-]+)\.([a-z\.]{2,6})([\/\w \.-]*)*\/?$/);
        default:
          return this.pattern;
      }
    }
  },
  mounted() {
    // Create a <style> element
    const style = document.createElement('style');

    // Define the CSS rules as a string
    style.innerHTML = `
      .entries-input {
        display: flex;
        flex-wrap: wrap;
        align-items: center;
        gap: 0.25rem;
        padding: 0.5rem;
        border: 1px solid #dbdbdb;
        border-radius: 4px;
      }
      .entries-input .input {
        flex-grow: 1;
        border: none;
        box-shadow: none;
        padding: 0;
        margin: 0;
      }
      .entries-input .input:focus {
        outline: none;
        box-shadow: none;
      }`

    // Append the <style> element to the <head> of the document
    document.head.appendChild(style);
  }
}
