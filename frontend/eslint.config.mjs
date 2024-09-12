import compat from "eslint-plugin-compat";
import globals from "globals";

export default [
  {
    ignores: ["**/*.ts"],
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
