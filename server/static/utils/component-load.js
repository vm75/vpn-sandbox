(function (root, factory) {
  if (typeof define === 'function' && define.amd) {
    define([], factory);
  } else if (typeof module === 'object' && module.exports) {
    module.exports = factory();
  } else {
    root.Component = factory();
  }
}(typeof self !== 'undefined' ? self : this, function () {

  var vueModules = {};
  var mountedComponents = {};

  function injectComponent({ name, source, elementId, data = {}, methods = {}, ref = '', parentElementId = null, onMount = null }) {
    var template = ``;

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
      components[name] = Vue.defineAsyncComponent(() => importComponent(source));
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

  // Utility function to download content from a URL, transpose it, create a Blob URL, and import the module
  async function __importVue(url) {
    try {
      // Step 1: Download the content of the URL
      const response = await fetch(url);
      const content = await response.text();

      // Step 2: Transpose the string content (You can modify this method as needed)
      const templateMatch = content.match(/<template>([\s\S]*?)<\/template>/);
      const styleMatch = content.match(/<style(?:[^>]*)>([\s\S]*?)<\/style>/);
      const scriptMatch = content.match(/<script>([\s\S]*)export\s+default\s*{([\s\S]*)}\s*<\/script>/);

      const template = templateMatch ? templateMatch[1].trim() : '';
      const imports = scriptMatch ? scriptMatch[1].trim() : '';
      const script = scriptMatch ? scriptMatch[2].trim() : '';
      const style = styleMatch ? styleMatch[1].trim() : '';

      injectStyle(url, style);

      const vueAsJs = `
        ${imports}
        export default {
          template: \`${template}\`,
          ${script}
        }`;

      // Step 3: Create a Blob from the transposed content
      const blob = new Blob([vueAsJs], { type: 'application/javascript' });

      // Step 4: Create a URL for the Blob
      const blobUrl = URL.createObjectURL(blob);

      // Step 5: Dynamically import the Blob URL
      return await import(blobUrl);
    } catch (error) {
      console.error('Error during download and import:', error);
    }
  }

  // Utility function to download content from a URL, transpose it, create a Blob URL, and import the module
  function importComponent(url) {
    if (!vueModules[url]) {
      if (url.endsWith('.js')) {
        url = window.location.origin + "/" + url;
        vueModules[url] = import(url)
      } else {
        if (!url.endsWith('.vue')) {
          url += '.vue';
        }
        vueModules[url] = __importVue(url);
      }
    }
    return vueModules[url];
  }

  // Return an object exposing the API
  return {
    isLoaded: isLoaded,
    inject: injectComponent,
    import: importComponent,
    injectStyleUrl: injectStyleUrl,
    injectScriptUrl: injectScriptUrl
  };
}));
