
var mountedComponents = {};

function injectComponent({ name, source, elementId, data = {}, methods = {}, ref = '', parentElementId = null, onMount = null }) {
  var template = ``;

  const getSourceUrl = (url) => {
    try {
      return new URL(url);
    } catch (e) {
      return new URL(url, document.baseURI);
    }
  }

  if (mountedComponents[elementId]) {
    mountedComponents[elementId].unmount();
  }

  template += `<${name} `;
  if (ref) {
    template += `ref="${ref}" `;
  }
  for (var key in data) {
    if (typeof data[key] === 'function') {
      template += `:${key}="${key}" `;
    } else {
      template += `v-model:${key}="${key}" `;
    }
  }
  for (var methodName in methods) {
    if (methodName.endsWith('OnUpdate')) {
      var dataKey = methodName.replace('OnUpdate', '');
      template += `@update:${dataKey}="${methodName}" `;
    } else {
      template += `@${methodName}="${methodName}" `;
    }
  }
  template += `></${name}>`;

  var components = {};

  if (typeof source === 'string') { // if component is a string, it's a URL
    const sourceUrl = getSourceUrl(source);
    components[name] = Vue.defineAsyncComponent(() => import(sourceUrl));
  } else {
    components[name] = source;
  }

  const app = Vue.createApp({
    components: components,
    data() {
      return data;
    },
    methods: methods,
    template: template,
    mounted() {
      if (onMount) {
        onMount();
      }
    },
  });

  // if targetElementId does not exist, create it
  if (!document.getElementById(elementId)) {
    var parent = null;
    if (parentElementId) {
      parent = document.getElementById(parentElementId);
    }
    if (!parent) {
      parent = document.body;
    }
    parent.insertAdjacentHTML('beforeend', `<div id="${elementId}"></div>`);
  }

  // Mount the dynamic app to the specified target DOM element
  app.mount(`#${elementId}`);

  mountedComponents[elementId] = app;

  return app._instance;
}

function injectStyle(id, css) {
  if (isLoaded(id)) {
    return false;
  }

  // Create a <style> element
  const style = document.createElement('style');
  style.id = id;
  style.textContent = css

  // Append the <style> element to the <head> of the document
  document.head.appendChild(style);
}


function injectStyleUrl(url) {
  if (isLoaded(url)) {
    return false;
  }
  // Create a <link> element
  const link = document.createElement('link');
  link.id = url;
  link.rel = 'stylesheet';
  link.href = url;
  document.head.appendChild(link);

  return true;
}

function injectScriptUrl(url, onload = null) {
  if (isLoaded(url)) {
    return false;
  }
  // Create a <script> element
  const script = document.createElement('script');
  script.id = url;
  script.src = url;
  if (onload) {
    script.onload = onload;
  }
  document.head.appendChild(script);

  return true;
}

function isLoaded(url) {
  return !!document.getElementById(url);
}
