import compat from "eslint-plugin-compat";
import globals from "globals";

export default [
  {
    plugins: {
      compat,
    },

    languageOptions: {
      globals: {
        ...globals.browser,
      },
    },

    rules: {
      "compat/compat": "error",
    },
  },
];
