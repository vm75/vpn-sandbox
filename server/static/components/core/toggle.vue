<template>
  <div>
    <input v-if="type === 'checkbox'" :id="id" type="checkbox" v-model="internalValue" @change="emitInput" />
    <label v-if="type === 'switch'" class="toggle-switch">
      <input :id="id" type="checkbox" v-model="internalValue" @change="emitInput" />
      <span class="toggle-slider round"></span>
    </label>
  </div>
</template>

<script>
// Toggle Input Component
export default {
  name: 'toggle',
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
      type: Boolean,
      default: false
    },
  },
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
  }
}
</script>

<style>
/* The switch - the box around the slider */
.toggle-switch {
  position: relative;
  display: inline-block;
  width: 60px;
  height: 34px;
}

/* Hide default HTML checkbox */
.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

/* The slider */
.toggle-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #ccc;
  -webkit-transition: .4s;
  transition: .4s;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 26px;
  width: 26px;
  left: 4px;
  bottom: 4px;
  background-color: white;
  -webkit-transition: .4s;
  transition: .4s;
}

input:checked+.toggle-slider {
  background-color: #2196F3;
}

input:focus+.toggle-slider {
  box-shadow: 0 0 1px #2196F3;
}

input:checked+.toggle-slider:before {
  -webkit-transform: translateX(26px);
  -ms-transform: translateX(26px);
  transform: translateX(26px);
}

/* Rounded sliders */
.toggle-slider.round {
  border-radius: 34px;
}

.toggle-slider.round:before {
  border-radius: 50%;
}
</style>
