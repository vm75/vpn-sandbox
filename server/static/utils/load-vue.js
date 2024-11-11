function convertVueToJs(content) {
    const templateMatch = content.match(/<template>([\s\S]*?)<\/template>/);
    const styleMatch = content.match(/<style(?:[^>]*)>([\s\S]*?)<\/style>/);
    const scriptMatch = content.match(/<script>\s*export\s+default\s*{([\s\S]*?)}\s*<\/script>/);

    const template = templateMatch ? templateMatch[1].trim() : '';
    const script = scriptMatch ? scriptMatch[1].trim() : '';
    const style = styleMatch ? styleMatch[1].trim() : '';

    return `export default {
    template: \`${template}\`,
    style: \`${style}\`,
    ${script}
  }`;
}

// Utility function to download content from a URL, transpose it, create a Blob URL, and import the module
async function loadVue(url) {
    try {
        // Step 1: Download the content of the URL
        const response = await fetch(url);
        const content = await response.text();

        // Step 2: Transpose the string content (You can modify this method as needed)
        const vueAsJs = convertVueToJs(content);

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
