/** @type {import('postcss-load-config').Config} */
const config = {
  plugins: [
    require("autoprefixer")({
      flexbox: true,
      overrideBrowserslist: ["> 0.01%"],
    }),
  ],
};

module.exports = config;
