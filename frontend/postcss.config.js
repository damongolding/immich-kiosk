/** @type {import('postcss-load-config').Config} */
const config = {
    plugins: [
        require("autoprefixer")({
            overrideBrowserslist: ["> 0.01%"],
        }),
        require("postcss-nested"),
    ],
};

module.exports = config;
