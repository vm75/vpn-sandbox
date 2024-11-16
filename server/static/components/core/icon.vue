<template>
  <div class="icon-container">
    <!-- Main Icon -->
    <img :src="icon" alt="Main Icon" class="main-icon">

    <!-- Banner Icon -->
    <img v-if="banner !== null" :src="banner" class="banner-icon">
  </div>
</template>

<script>
export default {
  props: {
    icon: {
      type: String,
      required: true
    },
    banner: {
      type: String,
      default: null
    }
  },
  watch: {
    icon() {
      this.$forceUpdate();
    },
    banner() {
      this.$forceUpdate();
    }
  }
}
</script>

<style scoped>
.icon-container {
  position: relative;
  width: 2em;
  height: 2em;
}

.main-icon {
  width: 100%;
  height: 100%;
}

.banner-icon {
  position: absolute;
  top: -0.5em;
  /* Adjust as needed */
  right: -0.5em;
  /* Adjust as needed */
  width: 1em;
  height: 1em;
}
</style>
