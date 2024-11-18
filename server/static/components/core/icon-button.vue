<template>
  <a href="{{ href }}" target="_blank" :class="`button ${classx}`">
    <span class="icon">
      <img :src="icon" alt="{{ text }}" :style="`width: ${size}em; height: ${size}em;`">
    </span>
    <span>{{ text }}</span>
  </a>
</template>

<script>
// Icon Button Component
export default {
  name: "icon-button",
  props: {
    icon: {
      type: String,
      required: true
    },
    href: {
      type: String,
      default: ''
    },
    text: {
      type: String,
      default: ''
    },
    size: {
      type: Number,
      default: 1
    },
    classx: {
      type: String,
      default: ''
    },
  },
}
</script>

<style></style>
