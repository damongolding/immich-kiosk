"use strict";
var __getOwnPropNames2 = Object.getOwnPropertyNames;
var __commonJS = function(cb, mod) {
  return function __require() {
    return mod || (0, cb[__getOwnPropNames2(cb)[0]])((mod = { exports: {} }).exports, mod), mod.exports;
  };
};

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/global-this.js
var require_global_this = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/global-this.js": function(exports2, module2) {
    "use strict";
    var check = function(it) {
      return it && it.Math === Math && it;
    };
    module2.exports = // eslint-disable-next-line es/no-global-this -- safe
    check(typeof globalThis == "object" && globalThis) || check(typeof window == "object" && window) || // eslint-disable-next-line no-restricted-globals -- safe
    check(typeof self == "object" && self) || check(typeof global == "object" && global) || check(typeof exports2 == "object" && exports2) || // eslint-disable-next-line no-new-func -- fallback
    /* @__PURE__ */ function() {
      return this;
    }() || Function("return this")();
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/path.js
var require_path = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/path.js": function(exports2, module2) {
    "use strict";
    var globalThis2 = require_global_this();
    module2.exports = globalThis2;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/fails.js
var require_fails = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/fails.js": function(exports2, module2) {
    "use strict";
    module2.exports = function(exec) {
      try {
        return !!exec();
      } catch (error) {
        return true;
      }
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/function-bind-native.js
var require_function_bind_native = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/function-bind-native.js": function(exports2, module2) {
    "use strict";
    var fails = require_fails();
    module2.exports = !fails(function() {
      var test = function() {
      }.bind();
      return typeof test != "function" || test.hasOwnProperty("prototype");
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/function-uncurry-this.js
var require_function_uncurry_this = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/function-uncurry-this.js": function(exports2, module2) {
    "use strict";
    var NATIVE_BIND = require_function_bind_native();
    var FunctionPrototype = Function.prototype;
    var call = FunctionPrototype.call;
    var uncurryThisWithBind = NATIVE_BIND && FunctionPrototype.bind.bind(call, call);
    module2.exports = NATIVE_BIND ? uncurryThisWithBind : function(fn) {
      return function() {
        return call.apply(fn, arguments);
      };
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-null-or-undefined.js
var require_is_null_or_undefined = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-null-or-undefined.js": function(exports2, module2) {
    "use strict";
    module2.exports = function(it) {
      return it === null || it === void 0;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/require-object-coercible.js
var require_require_object_coercible = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/require-object-coercible.js": function(exports2, module2) {
    "use strict";
    var isNullOrUndefined = require_is_null_or_undefined();
    var $TypeError = TypeError;
    module2.exports = function(it) {
      if (isNullOrUndefined(it)) throw new $TypeError("Can't call method on " + it);
      return it;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/to-object.js
var require_to_object = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/to-object.js": function(exports2, module2) {
    "use strict";
    var requireObjectCoercible = require_require_object_coercible();
    var $Object = Object;
    module2.exports = function(argument) {
      return $Object(requireObjectCoercible(argument));
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/has-own-property.js
var require_has_own_property = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/has-own-property.js": function(exports2, module2) {
    "use strict";
    var uncurryThis = require_function_uncurry_this();
    var toObject = require_to_object();
    var hasOwnProperty = uncurryThis({}.hasOwnProperty);
    module2.exports = Object.hasOwn || function hasOwn(it, key) {
      return hasOwnProperty(toObject(it), key);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-pure.js
var require_is_pure = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-pure.js": function(exports2, module2) {
    "use strict";
    module2.exports = false;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/define-global-property.js
var require_define_global_property = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/define-global-property.js": function(exports2, module2) {
    "use strict";
    var globalThis2 = require_global_this();
    var defineProperty = Object.defineProperty;
    module2.exports = function(key, value) {
      try {
        defineProperty(globalThis2, key, { value: value, configurable: true, writable: true });
      } catch (error) {
        globalThis2[key] = value;
      }
      return value;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/shared-store.js
var require_shared_store = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/shared-store.js": function(exports2, module2) {
    "use strict";
    var IS_PURE = require_is_pure();
    var globalThis2 = require_global_this();
    var defineGlobalProperty = require_define_global_property();
    var SHARED = "__core-js_shared__";
    var store = module2.exports = globalThis2[SHARED] || defineGlobalProperty(SHARED, {});
    (store.versions || (store.versions = [])).push({
      version: "3.39.0",
      mode: IS_PURE ? "pure" : "global",
      copyright: "\xA9 2014-2024 Denis Pushkarev (zloirock.ru)",
      license: "https://github.com/zloirock/core-js/blob/v3.39.0/LICENSE",
      source: "https://github.com/zloirock/core-js"
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/shared.js
var require_shared = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/shared.js": function(exports2, module2) {
    "use strict";
    var store = require_shared_store();
    module2.exports = function(key, value) {
      return store[key] || (store[key] = value || {});
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/uid.js
var require_uid = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/uid.js": function(exports2, module2) {
    "use strict";
    var uncurryThis = require_function_uncurry_this();
    var id = 0;
    var postfix = Math.random();
    var toString = uncurryThis(1 .toString);
    module2.exports = function(key) {
      return "Symbol(" + (key === void 0 ? "" : key) + ")_" + toString(++id + postfix, 36);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/environment-user-agent.js
var require_environment_user_agent = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/environment-user-agent.js": function(exports2, module2) {
    "use strict";
    var globalThis2 = require_global_this();
    var navigator2 = globalThis2.navigator;
    var userAgent = navigator2 && navigator2.userAgent;
    module2.exports = userAgent ? String(userAgent) : "";
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/environment-v8-version.js
var require_environment_v8_version = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/environment-v8-version.js": function(exports2, module2) {
    "use strict";
    var globalThis2 = require_global_this();
    var userAgent = require_environment_user_agent();
    var process = globalThis2.process;
    var Deno2 = globalThis2.Deno;
    var versions = process && process.versions || Deno2 && Deno2.version;
    var v8 = versions && versions.v8;
    var match;
    var version;
    if (v8) {
      match = v8.split(".");
      version = match[0] > 0 && match[0] < 4 ? 1 : +(match[0] + match[1]);
    }
    if (!version && userAgent) {
      match = userAgent.match(/Edge\/(\d+)/);
      if (!match || match[1] >= 74) {
        match = userAgent.match(/Chrome\/(\d+)/);
        if (match) version = +match[1];
      }
    }
    module2.exports = version;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/symbol-constructor-detection.js
var require_symbol_constructor_detection = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/symbol-constructor-detection.js": function(exports2, module2) {
    "use strict";
    var V8_VERSION = require_environment_v8_version();
    var fails = require_fails();
    var globalThis2 = require_global_this();
    var $String = globalThis2.String;
    module2.exports = !!Object.getOwnPropertySymbols && !fails(function() {
      var symbol = Symbol("symbol detection");
      return !$String(symbol) || !(Object(symbol) instanceof Symbol) || // Chrome 38-40 symbols are not inherited from DOM collections prototypes to instances
      !Symbol.sham && V8_VERSION && V8_VERSION < 41;
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/use-symbol-as-uid.js
var require_use_symbol_as_uid = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/use-symbol-as-uid.js": function(exports2, module2) {
    "use strict";
    var NATIVE_SYMBOL = require_symbol_constructor_detection();
    module2.exports = NATIVE_SYMBOL && !Symbol.sham && typeof Symbol.iterator == "symbol";
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/well-known-symbol.js
var require_well_known_symbol = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/well-known-symbol.js": function(exports2, module2) {
    "use strict";
    var globalThis2 = require_global_this();
    var shared = require_shared();
    var hasOwn = require_has_own_property();
    var uid = require_uid();
    var NATIVE_SYMBOL = require_symbol_constructor_detection();
    var USE_SYMBOL_AS_UID = require_use_symbol_as_uid();
    var Symbol2 = globalThis2.Symbol;
    var WellKnownSymbolsStore = shared("wks");
    var createWellKnownSymbol = USE_SYMBOL_AS_UID ? Symbol2["for"] || Symbol2 : Symbol2 && Symbol2.withoutSetter || uid;
    module2.exports = function(name) {
      if (!hasOwn(WellKnownSymbolsStore, name)) {
        WellKnownSymbolsStore[name] = NATIVE_SYMBOL && hasOwn(Symbol2, name) ? Symbol2[name] : createWellKnownSymbol("Symbol." + name);
      }
      return WellKnownSymbolsStore[name];
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/well-known-symbol-wrapped.js
var require_well_known_symbol_wrapped = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/well-known-symbol-wrapped.js": function(exports2) {
    "use strict";
    var wellKnownSymbol = require_well_known_symbol();
    exports2.f = wellKnownSymbol;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/descriptors.js
var require_descriptors = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/descriptors.js": function(exports2, module2) {
    "use strict";
    var fails = require_fails();
    module2.exports = !fails(function() {
      return Object.defineProperty({}, 1, { get: function() {
        return 7;
      } })[1] !== 7;
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-callable.js
var require_is_callable = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-callable.js": function(exports2, module2) {
    "use strict";
    var documentAll = typeof document == "object" && document.all;
    module2.exports = typeof documentAll == "undefined" && documentAll !== void 0 ? function(argument) {
      return typeof argument == "function" || argument === documentAll;
    } : function(argument) {
      return typeof argument == "function";
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-object.js
var require_is_object = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-object.js": function(exports2, module2) {
    "use strict";
    var isCallable = require_is_callable();
    module2.exports = function(it) {
      return typeof it == "object" ? it !== null : isCallable(it);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/document-create-element.js
var require_document_create_element = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/document-create-element.js": function(exports2, module2) {
    "use strict";
    var globalThis2 = require_global_this();
    var isObject = require_is_object();
    var document2 = globalThis2.document;
    var EXISTS = isObject(document2) && isObject(document2.createElement);
    module2.exports = function(it) {
      return EXISTS ? document2.createElement(it) : {};
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/ie8-dom-define.js
var require_ie8_dom_define = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/ie8-dom-define.js": function(exports2, module2) {
    "use strict";
    var DESCRIPTORS = require_descriptors();
    var fails = require_fails();
    var createElement = require_document_create_element();
    module2.exports = !DESCRIPTORS && !fails(function() {
      return Object.defineProperty(createElement("div"), "a", {
        get: function() {
          return 7;
        }
      }).a !== 7;
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/v8-prototype-define-bug.js
var require_v8_prototype_define_bug = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/v8-prototype-define-bug.js": function(exports2, module2) {
    "use strict";
    var DESCRIPTORS = require_descriptors();
    var fails = require_fails();
    module2.exports = DESCRIPTORS && fails(function() {
      return Object.defineProperty(function() {
      }, "prototype", {
        value: 42,
        writable: false
      }).prototype !== 42;
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/an-object.js
var require_an_object = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/an-object.js": function(exports2, module2) {
    "use strict";
    var isObject = require_is_object();
    var $String = String;
    var $TypeError = TypeError;
    module2.exports = function(argument) {
      if (isObject(argument)) return argument;
      throw new $TypeError($String(argument) + " is not an object");
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/function-call.js
var require_function_call = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/function-call.js": function(exports2, module2) {
    "use strict";
    var NATIVE_BIND = require_function_bind_native();
    var call = Function.prototype.call;
    module2.exports = NATIVE_BIND ? call.bind(call) : function() {
      return call.apply(call, arguments);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/get-built-in.js
var require_get_built_in = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/get-built-in.js": function(exports2, module2) {
    "use strict";
    var globalThis2 = require_global_this();
    var isCallable = require_is_callable();
    var aFunction = function(argument) {
      return isCallable(argument) ? argument : void 0;
    };
    module2.exports = function(namespace, method) {
      return arguments.length < 2 ? aFunction(globalThis2[namespace]) : globalThis2[namespace] && globalThis2[namespace][method];
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-is-prototype-of.js
var require_object_is_prototype_of = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-is-prototype-of.js": function(exports2, module2) {
    "use strict";
    var uncurryThis = require_function_uncurry_this();
    module2.exports = uncurryThis({}.isPrototypeOf);
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-symbol.js
var require_is_symbol = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-symbol.js": function(exports2, module2) {
    "use strict";
    var getBuiltIn = require_get_built_in();
    var isCallable = require_is_callable();
    var isPrototypeOf = require_object_is_prototype_of();
    var USE_SYMBOL_AS_UID = require_use_symbol_as_uid();
    var $Object = Object;
    module2.exports = USE_SYMBOL_AS_UID ? function(it) {
      return typeof it == "symbol";
    } : function(it) {
      var $Symbol = getBuiltIn("Symbol");
      return isCallable($Symbol) && isPrototypeOf($Symbol.prototype, $Object(it));
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/try-to-string.js
var require_try_to_string = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/try-to-string.js": function(exports2, module2) {
    "use strict";
    var $String = String;
    module2.exports = function(argument) {
      try {
        return $String(argument);
      } catch (error) {
        return "Object";
      }
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/a-callable.js
var require_a_callable = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/a-callable.js": function(exports2, module2) {
    "use strict";
    var isCallable = require_is_callable();
    var tryToString = require_try_to_string();
    var $TypeError = TypeError;
    module2.exports = function(argument) {
      if (isCallable(argument)) return argument;
      throw new $TypeError(tryToString(argument) + " is not a function");
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/get-method.js
var require_get_method = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/get-method.js": function(exports2, module2) {
    "use strict";
    var aCallable = require_a_callable();
    var isNullOrUndefined = require_is_null_or_undefined();
    module2.exports = function(V, P) {
      var func = V[P];
      return isNullOrUndefined(func) ? void 0 : aCallable(func);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/ordinary-to-primitive.js
var require_ordinary_to_primitive = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/ordinary-to-primitive.js": function(exports2, module2) {
    "use strict";
    var call = require_function_call();
    var isCallable = require_is_callable();
    var isObject = require_is_object();
    var $TypeError = TypeError;
    module2.exports = function(input, pref) {
      var fn, val;
      if (pref === "string" && isCallable(fn = input.toString) && !isObject(val = call(fn, input))) return val;
      if (isCallable(fn = input.valueOf) && !isObject(val = call(fn, input))) return val;
      if (pref !== "string" && isCallable(fn = input.toString) && !isObject(val = call(fn, input))) return val;
      throw new $TypeError("Can't convert object to primitive value");
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/to-primitive.js
var require_to_primitive = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/to-primitive.js": function(exports2, module2) {
    "use strict";
    var call = require_function_call();
    var isObject = require_is_object();
    var isSymbol = require_is_symbol();
    var getMethod = require_get_method();
    var ordinaryToPrimitive = require_ordinary_to_primitive();
    var wellKnownSymbol = require_well_known_symbol();
    var $TypeError = TypeError;
    var TO_PRIMITIVE = wellKnownSymbol("toPrimitive");
    module2.exports = function(input, pref) {
      if (!isObject(input) || isSymbol(input)) return input;
      var exoticToPrim = getMethod(input, TO_PRIMITIVE);
      var result;
      if (exoticToPrim) {
        if (pref === void 0) pref = "default";
        result = call(exoticToPrim, input, pref);
        if (!isObject(result) || isSymbol(result)) return result;
        throw new $TypeError("Can't convert object to primitive value");
      }
      if (pref === void 0) pref = "number";
      return ordinaryToPrimitive(input, pref);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/to-property-key.js
var require_to_property_key = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/to-property-key.js": function(exports2, module2) {
    "use strict";
    var toPrimitive = require_to_primitive();
    var isSymbol = require_is_symbol();
    module2.exports = function(argument) {
      var key = toPrimitive(argument, "string");
      return isSymbol(key) ? key : key + "";
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-define-property.js
var require_object_define_property = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-define-property.js": function(exports2) {
    "use strict";
    var DESCRIPTORS = require_descriptors();
    var IE8_DOM_DEFINE = require_ie8_dom_define();
    var V8_PROTOTYPE_DEFINE_BUG = require_v8_prototype_define_bug();
    var anObject = require_an_object();
    var toPropertyKey = require_to_property_key();
    var $TypeError = TypeError;
    var $defineProperty = Object.defineProperty;
    var $getOwnPropertyDescriptor = Object.getOwnPropertyDescriptor;
    var ENUMERABLE = "enumerable";
    var CONFIGURABLE = "configurable";
    var WRITABLE = "writable";
    exports2.f = DESCRIPTORS ? V8_PROTOTYPE_DEFINE_BUG ? function defineProperty(O, P, Attributes) {
      anObject(O);
      P = toPropertyKey(P);
      anObject(Attributes);
      if (typeof O === "function" && P === "prototype" && "value" in Attributes && WRITABLE in Attributes && !Attributes[WRITABLE]) {
        var current = $getOwnPropertyDescriptor(O, P);
        if (current && current[WRITABLE]) {
          O[P] = Attributes.value;
          Attributes = {
            configurable: CONFIGURABLE in Attributes ? Attributes[CONFIGURABLE] : current[CONFIGURABLE],
            enumerable: ENUMERABLE in Attributes ? Attributes[ENUMERABLE] : current[ENUMERABLE],
            writable: false
          };
        }
      }
      return $defineProperty(O, P, Attributes);
    } : $defineProperty : function defineProperty(O, P, Attributes) {
      anObject(O);
      P = toPropertyKey(P);
      anObject(Attributes);
      if (IE8_DOM_DEFINE) try {
        return $defineProperty(O, P, Attributes);
      } catch (error) {
      }
      if ("get" in Attributes || "set" in Attributes) throw new $TypeError("Accessors not supported");
      if ("value" in Attributes) O[P] = Attributes.value;
      return O;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/well-known-symbol-define.js
var require_well_known_symbol_define = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/well-known-symbol-define.js": function(exports2, module2) {
    "use strict";
    var path = require_path();
    var hasOwn = require_has_own_property();
    var wrappedWellKnownSymbolModule = require_well_known_symbol_wrapped();
    var defineProperty = require_object_define_property().f;
    module2.exports = function(NAME) {
      var Symbol2 = path.Symbol || (path.Symbol = {});
      if (!hasOwn(Symbol2, NAME)) defineProperty(Symbol2, NAME, {
        value: wrappedWellKnownSymbolModule.f(NAME)
      });
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.symbol.async-iterator.js
var require_es_symbol_async_iterator = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.symbol.async-iterator.js": function() {
    "use strict";
    var defineWellKnownSymbol = require_well_known_symbol_define();
    defineWellKnownSymbol("asyncIterator");
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-property-is-enumerable.js
var require_object_property_is_enumerable = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-property-is-enumerable.js": function(exports2) {
    "use strict";
    var $propertyIsEnumerable = {}.propertyIsEnumerable;
    var getOwnPropertyDescriptor = Object.getOwnPropertyDescriptor;
    var NASHORN_BUG = getOwnPropertyDescriptor && !$propertyIsEnumerable.call({ 1: 2 }, 1);
    exports2.f = NASHORN_BUG ? function propertyIsEnumerable(V) {
      var descriptor = getOwnPropertyDescriptor(this, V);
      return !!descriptor && descriptor.enumerable;
    } : $propertyIsEnumerable;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/create-property-descriptor.js
var require_create_property_descriptor = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/create-property-descriptor.js": function(exports2, module2) {
    "use strict";
    module2.exports = function(bitmap, value) {
      return {
        enumerable: !(bitmap & 1),
        configurable: !(bitmap & 2),
        writable: !(bitmap & 4),
        value: value
      };
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/classof-raw.js
var require_classof_raw = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/classof-raw.js": function(exports2, module2) {
    "use strict";
    var uncurryThis = require_function_uncurry_this();
    var toString = uncurryThis({}.toString);
    var stringSlice = uncurryThis("".slice);
    module2.exports = function(it) {
      return stringSlice(toString(it), 8, -1);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/indexed-object.js
var require_indexed_object = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/indexed-object.js": function(exports2, module2) {
    "use strict";
    var uncurryThis = require_function_uncurry_this();
    var fails = require_fails();
    var classof = require_classof_raw();
    var $Object = Object;
    var split = uncurryThis("".split);
    module2.exports = fails(function() {
      return !$Object("z").propertyIsEnumerable(0);
    }) ? function(it) {
      return classof(it) === "String" ? split(it, "") : $Object(it);
    } : $Object;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/to-indexed-object.js
var require_to_indexed_object = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/to-indexed-object.js": function(exports2, module2) {
    "use strict";
    var IndexedObject = require_indexed_object();
    var requireObjectCoercible = require_require_object_coercible();
    module2.exports = function(it) {
      return IndexedObject(requireObjectCoercible(it));
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-get-own-property-descriptor.js
var require_object_get_own_property_descriptor = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-get-own-property-descriptor.js": function(exports2) {
    "use strict";
    var DESCRIPTORS = require_descriptors();
    var call = require_function_call();
    var propertyIsEnumerableModule = require_object_property_is_enumerable();
    var createPropertyDescriptor = require_create_property_descriptor();
    var toIndexedObject = require_to_indexed_object();
    var toPropertyKey = require_to_property_key();
    var hasOwn = require_has_own_property();
    var IE8_DOM_DEFINE = require_ie8_dom_define();
    var $getOwnPropertyDescriptor = Object.getOwnPropertyDescriptor;
    exports2.f = DESCRIPTORS ? $getOwnPropertyDescriptor : function getOwnPropertyDescriptor(O, P) {
      O = toIndexedObject(O);
      P = toPropertyKey(P);
      if (IE8_DOM_DEFINE) try {
        return $getOwnPropertyDescriptor(O, P);
      } catch (error) {
      }
      if (hasOwn(O, P)) return createPropertyDescriptor(!call(propertyIsEnumerableModule.f, O, P), O[P]);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/create-non-enumerable-property.js
var require_create_non_enumerable_property = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/create-non-enumerable-property.js": function(exports2, module2) {
    "use strict";
    var DESCRIPTORS = require_descriptors();
    var definePropertyModule = require_object_define_property();
    var createPropertyDescriptor = require_create_property_descriptor();
    module2.exports = DESCRIPTORS ? function(object, key, value) {
      return definePropertyModule.f(object, key, createPropertyDescriptor(1, value));
    } : function(object, key, value) {
      object[key] = value;
      return object;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/function-name.js
var require_function_name = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/function-name.js": function(exports2, module2) {
    "use strict";
    var DESCRIPTORS = require_descriptors();
    var hasOwn = require_has_own_property();
    var FunctionPrototype = Function.prototype;
    var getDescriptor = DESCRIPTORS && Object.getOwnPropertyDescriptor;
    var EXISTS = hasOwn(FunctionPrototype, "name");
    var PROPER = EXISTS && function something() {
    }.name === "something";
    var CONFIGURABLE = EXISTS && (!DESCRIPTORS || DESCRIPTORS && getDescriptor(FunctionPrototype, "name").configurable);
    module2.exports = {
      EXISTS: EXISTS,
      PROPER: PROPER,
      CONFIGURABLE: CONFIGURABLE
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/inspect-source.js
var require_inspect_source = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/inspect-source.js": function(exports2, module2) {
    "use strict";
    var uncurryThis = require_function_uncurry_this();
    var isCallable = require_is_callable();
    var store = require_shared_store();
    var functionToString = uncurryThis(Function.toString);
    if (!isCallable(store.inspectSource)) {
      store.inspectSource = function(it) {
        return functionToString(it);
      };
    }
    module2.exports = store.inspectSource;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/weak-map-basic-detection.js
var require_weak_map_basic_detection = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/weak-map-basic-detection.js": function(exports2, module2) {
    "use strict";
    var globalThis2 = require_global_this();
    var isCallable = require_is_callable();
    var WeakMap2 = globalThis2.WeakMap;
    module2.exports = isCallable(WeakMap2) && /native code/.test(String(WeakMap2));
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/shared-key.js
var require_shared_key = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/shared-key.js": function(exports2, module2) {
    "use strict";
    var shared = require_shared();
    var uid = require_uid();
    var keys = shared("keys");
    module2.exports = function(key) {
      return keys[key] || (keys[key] = uid(key));
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/hidden-keys.js
var require_hidden_keys = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/hidden-keys.js": function(exports2, module2) {
    "use strict";
    module2.exports = {};
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/internal-state.js
var require_internal_state = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/internal-state.js": function(exports2, module2) {
    "use strict";
    var NATIVE_WEAK_MAP = require_weak_map_basic_detection();
    var globalThis2 = require_global_this();
    var isObject = require_is_object();
    var createNonEnumerableProperty = require_create_non_enumerable_property();
    var hasOwn = require_has_own_property();
    var shared = require_shared_store();
    var sharedKey = require_shared_key();
    var hiddenKeys = require_hidden_keys();
    var OBJECT_ALREADY_INITIALIZED = "Object already initialized";
    var TypeError2 = globalThis2.TypeError;
    var WeakMap2 = globalThis2.WeakMap;
    var set;
    var get;
    var has;
    var enforce = function(it) {
      return has(it) ? get(it) : set(it, {});
    };
    var getterFor = function(TYPE) {
      return function(it) {
        var state;
        if (!isObject(it) || (state = get(it)).type !== TYPE) {
          throw new TypeError2("Incompatible receiver, " + TYPE + " required");
        }
        return state;
      };
    };
    if (NATIVE_WEAK_MAP || shared.state) {
      store = shared.state || (shared.state = new WeakMap2());
      store.get = store.get;
      store.has = store.has;
      store.set = store.set;
      set = function(it, metadata) {
        if (store.has(it)) throw new TypeError2(OBJECT_ALREADY_INITIALIZED);
        metadata.facade = it;
        store.set(it, metadata);
        return metadata;
      };
      get = function(it) {
        return store.get(it) || {};
      };
      has = function(it) {
        return store.has(it);
      };
    } else {
      STATE = sharedKey("state");
      hiddenKeys[STATE] = true;
      set = function(it, metadata) {
        if (hasOwn(it, STATE)) throw new TypeError2(OBJECT_ALREADY_INITIALIZED);
        metadata.facade = it;
        createNonEnumerableProperty(it, STATE, metadata);
        return metadata;
      };
      get = function(it) {
        return hasOwn(it, STATE) ? it[STATE] : {};
      };
      has = function(it) {
        return hasOwn(it, STATE);
      };
    }
    var store;
    var STATE;
    module2.exports = {
      set: set,
      get: get,
      has: has,
      enforce: enforce,
      getterFor: getterFor
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/make-built-in.js
var require_make_built_in = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/make-built-in.js": function(exports2, module2) {
    "use strict";
    var uncurryThis = require_function_uncurry_this();
    var fails = require_fails();
    var isCallable = require_is_callable();
    var hasOwn = require_has_own_property();
    var DESCRIPTORS = require_descriptors();
    var CONFIGURABLE_FUNCTION_NAME = require_function_name().CONFIGURABLE;
    var inspectSource = require_inspect_source();
    var InternalStateModule = require_internal_state();
    var enforceInternalState = InternalStateModule.enforce;
    var getInternalState = InternalStateModule.get;
    var $String = String;
    var defineProperty = Object.defineProperty;
    var stringSlice = uncurryThis("".slice);
    var replace = uncurryThis("".replace);
    var join = uncurryThis([].join);
    var CONFIGURABLE_LENGTH = DESCRIPTORS && !fails(function() {
      return defineProperty(function() {
      }, "length", { value: 8 }).length !== 8;
    });
    var TEMPLATE = String(String).split("String");
    var makeBuiltIn = module2.exports = function(value, name, options) {
      if (stringSlice($String(name), 0, 7) === "Symbol(") {
        name = "[" + replace($String(name), /^Symbol\(([^)]*)\).*$/, "$1") + "]";
      }
      if (options && options.getter) name = "get " + name;
      if (options && options.setter) name = "set " + name;
      if (!hasOwn(value, "name") || CONFIGURABLE_FUNCTION_NAME && value.name !== name) {
        if (DESCRIPTORS) defineProperty(value, "name", { value: name, configurable: true });
        else value.name = name;
      }
      if (CONFIGURABLE_LENGTH && options && hasOwn(options, "arity") && value.length !== options.arity) {
        defineProperty(value, "length", { value: options.arity });
      }
      try {
        if (options && hasOwn(options, "constructor") && options.constructor) {
          if (DESCRIPTORS) defineProperty(value, "prototype", { writable: false });
        } else if (value.prototype) value.prototype = void 0;
      } catch (error) {
      }
      var state = enforceInternalState(value);
      if (!hasOwn(state, "source")) {
        state.source = join(TEMPLATE, typeof name == "string" ? name : "");
      }
      return value;
    };
    Function.prototype.toString = makeBuiltIn(function toString() {
      return isCallable(this) && getInternalState(this).source || inspectSource(this);
    }, "toString");
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/define-built-in.js
var require_define_built_in = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/define-built-in.js": function(exports2, module2) {
    "use strict";
    var isCallable = require_is_callable();
    var definePropertyModule = require_object_define_property();
    var makeBuiltIn = require_make_built_in();
    var defineGlobalProperty = require_define_global_property();
    module2.exports = function(O, key, value, options) {
      if (!options) options = {};
      var simple = options.enumerable;
      var name = options.name !== void 0 ? options.name : key;
      if (isCallable(value)) makeBuiltIn(value, name, options);
      if (options.global) {
        if (simple) O[key] = value;
        else defineGlobalProperty(key, value);
      } else {
        try {
          if (!options.unsafe) delete O[key];
          else if (O[key]) simple = true;
        } catch (error) {
        }
        if (simple) O[key] = value;
        else definePropertyModule.f(O, key, {
          value: value,
          enumerable: false,
          configurable: !options.nonConfigurable,
          writable: !options.nonWritable
        });
      }
      return O;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/math-trunc.js
var require_math_trunc = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/math-trunc.js": function(exports2, module2) {
    "use strict";
    var ceil = Math.ceil;
    var floor = Math.floor;
    module2.exports = Math.trunc || function trunc(x) {
      var n = +x;
      return (n > 0 ? floor : ceil)(n);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/to-integer-or-infinity.js
var require_to_integer_or_infinity = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/to-integer-or-infinity.js": function(exports2, module2) {
    "use strict";
    var trunc = require_math_trunc();
    module2.exports = function(argument) {
      var number = +argument;
      return number !== number || number === 0 ? 0 : trunc(number);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/to-absolute-index.js
var require_to_absolute_index = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/to-absolute-index.js": function(exports2, module2) {
    "use strict";
    var toIntegerOrInfinity = require_to_integer_or_infinity();
    var max = Math.max;
    var min = Math.min;
    module2.exports = function(index, length) {
      var integer = toIntegerOrInfinity(index);
      return integer < 0 ? max(integer + length, 0) : min(integer, length);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/to-length.js
var require_to_length = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/to-length.js": function(exports2, module2) {
    "use strict";
    var toIntegerOrInfinity = require_to_integer_or_infinity();
    var min = Math.min;
    module2.exports = function(argument) {
      var len = toIntegerOrInfinity(argument);
      return len > 0 ? min(len, 9007199254740991) : 0;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/length-of-array-like.js
var require_length_of_array_like = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/length-of-array-like.js": function(exports2, module2) {
    "use strict";
    var toLength = require_to_length();
    module2.exports = function(obj) {
      return toLength(obj.length);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/array-includes.js
var require_array_includes = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/array-includes.js": function(exports2, module2) {
    "use strict";
    var toIndexedObject = require_to_indexed_object();
    var toAbsoluteIndex = require_to_absolute_index();
    var lengthOfArrayLike = require_length_of_array_like();
    var createMethod = function(IS_INCLUDES) {
      return function($this, el, fromIndex) {
        var O = toIndexedObject($this);
        var length = lengthOfArrayLike(O);
        if (length === 0) return !IS_INCLUDES && -1;
        var index = toAbsoluteIndex(fromIndex, length);
        var value;
        if (IS_INCLUDES && el !== el) while (length > index) {
          value = O[index++];
          if (value !== value) return true;
        }
        else for (; length > index; index++) {
          if ((IS_INCLUDES || index in O) && O[index] === el) return IS_INCLUDES || index || 0;
        }
        return !IS_INCLUDES && -1;
      };
    };
    module2.exports = {
      // `Array.prototype.includes` method
      // https://tc39.es/ecma262/#sec-array.prototype.includes
      includes: createMethod(true),
      // `Array.prototype.indexOf` method
      // https://tc39.es/ecma262/#sec-array.prototype.indexof
      indexOf: createMethod(false)
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-keys-internal.js
var require_object_keys_internal = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-keys-internal.js": function(exports2, module2) {
    "use strict";
    var uncurryThis = require_function_uncurry_this();
    var hasOwn = require_has_own_property();
    var toIndexedObject = require_to_indexed_object();
    var indexOf = require_array_includes().indexOf;
    var hiddenKeys = require_hidden_keys();
    var push = uncurryThis([].push);
    module2.exports = function(object, names) {
      var O = toIndexedObject(object);
      var i = 0;
      var result = [];
      var key;
      for (key in O) !hasOwn(hiddenKeys, key) && hasOwn(O, key) && push(result, key);
      while (names.length > i) if (hasOwn(O, key = names[i++])) {
        ~indexOf(result, key) || push(result, key);
      }
      return result;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/enum-bug-keys.js
var require_enum_bug_keys = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/enum-bug-keys.js": function(exports2, module2) {
    "use strict";
    module2.exports = [
      "constructor",
      "hasOwnProperty",
      "isPrototypeOf",
      "propertyIsEnumerable",
      "toLocaleString",
      "toString",
      "valueOf"
    ];
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-get-own-property-names.js
var require_object_get_own_property_names = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-get-own-property-names.js": function(exports2) {
    "use strict";
    var internalObjectKeys = require_object_keys_internal();
    var enumBugKeys = require_enum_bug_keys();
    var hiddenKeys = enumBugKeys.concat("length", "prototype");
    exports2.f = Object.getOwnPropertyNames || function getOwnPropertyNames(O) {
      return internalObjectKeys(O, hiddenKeys);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-get-own-property-symbols.js
var require_object_get_own_property_symbols = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-get-own-property-symbols.js": function(exports2) {
    "use strict";
    exports2.f = Object.getOwnPropertySymbols;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/own-keys.js
var require_own_keys = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/own-keys.js": function(exports2, module2) {
    "use strict";
    var getBuiltIn = require_get_built_in();
    var uncurryThis = require_function_uncurry_this();
    var getOwnPropertyNamesModule = require_object_get_own_property_names();
    var getOwnPropertySymbolsModule = require_object_get_own_property_symbols();
    var anObject = require_an_object();
    var concat = uncurryThis([].concat);
    module2.exports = getBuiltIn("Reflect", "ownKeys") || function ownKeys(it) {
      var keys = getOwnPropertyNamesModule.f(anObject(it));
      var getOwnPropertySymbols = getOwnPropertySymbolsModule.f;
      return getOwnPropertySymbols ? concat(keys, getOwnPropertySymbols(it)) : keys;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/copy-constructor-properties.js
var require_copy_constructor_properties = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/copy-constructor-properties.js": function(exports2, module2) {
    "use strict";
    var hasOwn = require_has_own_property();
    var ownKeys = require_own_keys();
    var getOwnPropertyDescriptorModule = require_object_get_own_property_descriptor();
    var definePropertyModule = require_object_define_property();
    module2.exports = function(target, source, exceptions) {
      var keys = ownKeys(source);
      var defineProperty = definePropertyModule.f;
      var getOwnPropertyDescriptor = getOwnPropertyDescriptorModule.f;
      for (var i = 0; i < keys.length; i++) {
        var key = keys[i];
        if (!hasOwn(target, key) && !(exceptions && hasOwn(exceptions, key))) {
          defineProperty(target, key, getOwnPropertyDescriptor(source, key));
        }
      }
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-forced.js
var require_is_forced = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-forced.js": function(exports2, module2) {
    "use strict";
    var fails = require_fails();
    var isCallable = require_is_callable();
    var replacement = /#|\.prototype\./;
    var isForced = function(feature, detection) {
      var value = data[normalize(feature)];
      return value === POLYFILL ? true : value === NATIVE ? false : isCallable(detection) ? fails(detection) : !!detection;
    };
    var normalize = isForced.normalize = function(string) {
      return String(string).replace(replacement, ".").toLowerCase();
    };
    var data = isForced.data = {};
    var NATIVE = isForced.NATIVE = "N";
    var POLYFILL = isForced.POLYFILL = "P";
    module2.exports = isForced;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/export.js
var require_export = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/export.js": function(exports2, module2) {
    "use strict";
    var globalThis2 = require_global_this();
    var getOwnPropertyDescriptor = require_object_get_own_property_descriptor().f;
    var createNonEnumerableProperty = require_create_non_enumerable_property();
    var defineBuiltIn = require_define_built_in();
    var defineGlobalProperty = require_define_global_property();
    var copyConstructorProperties = require_copy_constructor_properties();
    var isForced = require_is_forced();
    module2.exports = function(options, source) {
      var TARGET = options.target;
      var GLOBAL = options.global;
      var STATIC = options.stat;
      var FORCED, target, key, targetProperty, sourceProperty, descriptor;
      if (GLOBAL) {
        target = globalThis2;
      } else if (STATIC) {
        target = globalThis2[TARGET] || defineGlobalProperty(TARGET, {});
      } else {
        target = globalThis2[TARGET] && globalThis2[TARGET].prototype;
      }
      if (target) for (key in source) {
        sourceProperty = source[key];
        if (options.dontCallGetSet) {
          descriptor = getOwnPropertyDescriptor(target, key);
          targetProperty = descriptor && descriptor.value;
        } else targetProperty = target[key];
        FORCED = isForced(GLOBAL ? key : TARGET + (STATIC ? "." : "#") + key, options.forced);
        if (!FORCED && targetProperty !== void 0) {
          if (typeof sourceProperty == typeof targetProperty) continue;
          copyConstructorProperties(sourceProperty, targetProperty);
        }
        if (options.sham || targetProperty && targetProperty.sham) {
          createNonEnumerableProperty(sourceProperty, "sham", true);
        }
        defineBuiltIn(target, key, sourceProperty, options);
      }
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-array.js
var require_is_array = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-array.js": function(exports2, module2) {
    "use strict";
    var classof = require_classof_raw();
    module2.exports = Array.isArray || function isArray(argument) {
      return classof(argument) === "Array";
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.array.reverse.js
var require_es_array_reverse = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.array.reverse.js": function() {
    "use strict";
    var $ = require_export();
    var uncurryThis = require_function_uncurry_this();
    var isArray = require_is_array();
    var nativeReverse = uncurryThis([].reverse);
    var test = [1, 2];
    $({ target: "Array", proto: true, forced: String(test) === String(test.reverse()) }, {
      reverse: function reverse() {
        if (isArray(this)) this.length = this.length;
        return nativeReverse(this);
      }
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/to-string-tag-support.js
var require_to_string_tag_support = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/to-string-tag-support.js": function(exports2, module2) {
    "use strict";
    var wellKnownSymbol = require_well_known_symbol();
    var TO_STRING_TAG = wellKnownSymbol("toStringTag");
    var test = {};
    test[TO_STRING_TAG] = "z";
    module2.exports = String(test) === "[object z]";
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/classof.js
var require_classof = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/classof.js": function(exports2, module2) {
    "use strict";
    var TO_STRING_TAG_SUPPORT = require_to_string_tag_support();
    var isCallable = require_is_callable();
    var classofRaw = require_classof_raw();
    var wellKnownSymbol = require_well_known_symbol();
    var TO_STRING_TAG = wellKnownSymbol("toStringTag");
    var $Object = Object;
    var CORRECT_ARGUMENTS = classofRaw(/* @__PURE__ */ function() {
      return arguments;
    }()) === "Arguments";
    var tryGet = function(it, key) {
      try {
        return it[key];
      } catch (error) {
      }
    };
    module2.exports = TO_STRING_TAG_SUPPORT ? classofRaw : function(it) {
      var O, tag, result;
      return it === void 0 ? "Undefined" : it === null ? "Null" : typeof (tag = tryGet(O = $Object(it), TO_STRING_TAG)) == "string" ? tag : CORRECT_ARGUMENTS ? classofRaw(O) : (result = classofRaw(O)) === "Object" && isCallable(O.callee) ? "Arguments" : result;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/to-string.js
var require_to_string = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/to-string.js": function(exports2, module2) {
    "use strict";
    var classof = require_classof();
    var $String = String;
    module2.exports = function(argument) {
      if (classof(argument) === "Symbol") throw new TypeError("Cannot convert a Symbol value to a string");
      return $String(argument);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/define-built-in-accessor.js
var require_define_built_in_accessor = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/define-built-in-accessor.js": function(exports2, module2) {
    "use strict";
    var makeBuiltIn = require_make_built_in();
    var defineProperty = require_object_define_property();
    module2.exports = function(target, name, descriptor) {
      if (descriptor.get) makeBuiltIn(descriptor.get, name, { getter: true });
      if (descriptor.set) makeBuiltIn(descriptor.set, name, { setter: true });
      return defineProperty.f(target, name, descriptor);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.symbol.description.js
var require_es_symbol_description = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.symbol.description.js": function() {
    "use strict";
    var $ = require_export();
    var DESCRIPTORS = require_descriptors();
    var globalThis2 = require_global_this();
    var uncurryThis = require_function_uncurry_this();
    var hasOwn = require_has_own_property();
    var isCallable = require_is_callable();
    var isPrototypeOf = require_object_is_prototype_of();
    var toString = require_to_string();
    var defineBuiltInAccessor = require_define_built_in_accessor();
    var copyConstructorProperties = require_copy_constructor_properties();
    var NativeSymbol = globalThis2.Symbol;
    var SymbolPrototype = NativeSymbol && NativeSymbol.prototype;
    if (DESCRIPTORS && isCallable(NativeSymbol) && (!("description" in SymbolPrototype) || // Safari 12 bug
    NativeSymbol().description !== void 0)) {
      EmptyStringDescriptionStore = {};
      SymbolWrapper = function Symbol2() {
        var description = arguments.length < 1 || arguments[0] === void 0 ? void 0 : toString(arguments[0]);
        var result = isPrototypeOf(SymbolPrototype, this) ? new NativeSymbol(description) : description === void 0 ? NativeSymbol() : NativeSymbol(description);
        if (description === "") EmptyStringDescriptionStore[result] = true;
        return result;
      };
      copyConstructorProperties(SymbolWrapper, NativeSymbol);
      SymbolWrapper.prototype = SymbolPrototype;
      SymbolPrototype.constructor = SymbolWrapper;
      NATIVE_SYMBOL = String(NativeSymbol("description detection")) === "Symbol(description detection)";
      thisSymbolValue = uncurryThis(SymbolPrototype.valueOf);
      symbolDescriptiveString = uncurryThis(SymbolPrototype.toString);
      regexp = /^Symbol\((.*)\)[^)]+$/;
      replace = uncurryThis("".replace);
      stringSlice = uncurryThis("".slice);
      defineBuiltInAccessor(SymbolPrototype, "description", {
        configurable: true,
        get: function description() {
          var symbol = thisSymbolValue(this);
          if (hasOwn(EmptyStringDescriptionStore, symbol)) return "";
          var string = symbolDescriptiveString(symbol);
          var desc = NATIVE_SYMBOL ? stringSlice(string, 7, -1) : replace(string, regexp, "$1");
          return desc === "" ? void 0 : desc;
        }
      });
      $({ global: true, constructor: true, forced: true }, {
        Symbol: SymbolWrapper
      });
    }
    var EmptyStringDescriptionStore;
    var SymbolWrapper;
    var NATIVE_SYMBOL;
    var thisSymbolValue;
    var symbolDescriptiveString;
    var regexp;
    var replace;
    var stringSlice;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/does-not-exceed-safe-integer.js
var require_does_not_exceed_safe_integer = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/does-not-exceed-safe-integer.js": function(exports2, module2) {
    "use strict";
    var $TypeError = TypeError;
    var MAX_SAFE_INTEGER = 9007199254740991;
    module2.exports = function(it) {
      if (it > MAX_SAFE_INTEGER) throw $TypeError("Maximum allowed index exceeded");
      return it;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/function-uncurry-this-clause.js
var require_function_uncurry_this_clause = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/function-uncurry-this-clause.js": function(exports2, module2) {
    "use strict";
    var classofRaw = require_classof_raw();
    var uncurryThis = require_function_uncurry_this();
    module2.exports = function(fn) {
      if (classofRaw(fn) === "Function") return uncurryThis(fn);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/function-bind-context.js
var require_function_bind_context = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/function-bind-context.js": function(exports2, module2) {
    "use strict";
    var uncurryThis = require_function_uncurry_this_clause();
    var aCallable = require_a_callable();
    var NATIVE_BIND = require_function_bind_native();
    var bind = uncurryThis(uncurryThis.bind);
    module2.exports = function(fn, that) {
      aCallable(fn);
      return that === void 0 ? fn : NATIVE_BIND ? bind(fn, that) : function() {
        return fn.apply(that, arguments);
      };
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/flatten-into-array.js
var require_flatten_into_array = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/flatten-into-array.js": function(exports2, module2) {
    "use strict";
    var isArray = require_is_array();
    var lengthOfArrayLike = require_length_of_array_like();
    var doesNotExceedSafeInteger = require_does_not_exceed_safe_integer();
    var bind = require_function_bind_context();
    var flattenIntoArray = function(target, original, source, sourceLen, start, depth, mapper, thisArg) {
      var targetIndex = start;
      var sourceIndex = 0;
      var mapFn = mapper ? bind(mapper, thisArg) : false;
      var element, elementLen;
      while (sourceIndex < sourceLen) {
        if (sourceIndex in source) {
          element = mapFn ? mapFn(source[sourceIndex], sourceIndex, original) : source[sourceIndex];
          if (depth > 0 && isArray(element)) {
            elementLen = lengthOfArrayLike(element);
            targetIndex = flattenIntoArray(target, original, element, elementLen, targetIndex, depth - 1) - 1;
          } else {
            doesNotExceedSafeInteger(targetIndex + 1);
            target[targetIndex] = element;
          }
          targetIndex++;
        }
        sourceIndex++;
      }
      return targetIndex;
    };
    module2.exports = flattenIntoArray;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-constructor.js
var require_is_constructor = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-constructor.js": function(exports2, module2) {
    "use strict";
    var uncurryThis = require_function_uncurry_this();
    var fails = require_fails();
    var isCallable = require_is_callable();
    var classof = require_classof();
    var getBuiltIn = require_get_built_in();
    var inspectSource = require_inspect_source();
    var noop = function() {
    };
    var construct = getBuiltIn("Reflect", "construct");
    var constructorRegExp = /^\s*(?:class|function)\b/;
    var exec = uncurryThis(constructorRegExp.exec);
    var INCORRECT_TO_STRING = !constructorRegExp.test(noop);
    var isConstructorModern = function isConstructor(argument) {
      if (!isCallable(argument)) return false;
      try {
        construct(noop, [], argument);
        return true;
      } catch (error) {
        return false;
      }
    };
    var isConstructorLegacy = function isConstructor(argument) {
      if (!isCallable(argument)) return false;
      switch (classof(argument)) {
        case "AsyncFunction":
        case "GeneratorFunction":
        case "AsyncGeneratorFunction":
          return false;
      }
      try {
        return INCORRECT_TO_STRING || !!exec(constructorRegExp, inspectSource(argument));
      } catch (error) {
        return true;
      }
    };
    isConstructorLegacy.sham = true;
    module2.exports = !construct || fails(function() {
      var called;
      return isConstructorModern(isConstructorModern.call) || !isConstructorModern(Object) || !isConstructorModern(function() {
        called = true;
      }) || called;
    }) ? isConstructorLegacy : isConstructorModern;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/array-species-constructor.js
var require_array_species_constructor = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/array-species-constructor.js": function(exports2, module2) {
    "use strict";
    var isArray = require_is_array();
    var isConstructor = require_is_constructor();
    var isObject = require_is_object();
    var wellKnownSymbol = require_well_known_symbol();
    var SPECIES = wellKnownSymbol("species");
    var $Array = Array;
    module2.exports = function(originalArray) {
      var C;
      if (isArray(originalArray)) {
        C = originalArray.constructor;
        if (isConstructor(C) && (C === $Array || isArray(C.prototype))) C = void 0;
        else if (isObject(C)) {
          C = C[SPECIES];
          if (C === null) C = void 0;
        }
      }
      return C === void 0 ? $Array : C;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/array-species-create.js
var require_array_species_create = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/array-species-create.js": function(exports2, module2) {
    "use strict";
    var arraySpeciesConstructor = require_array_species_constructor();
    module2.exports = function(originalArray, length) {
      return new (arraySpeciesConstructor(originalArray))(length === 0 ? 0 : length);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.array.flat.js
var require_es_array_flat = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.array.flat.js": function() {
    "use strict";
    var $ = require_export();
    var flattenIntoArray = require_flatten_into_array();
    var toObject = require_to_object();
    var lengthOfArrayLike = require_length_of_array_like();
    var toIntegerOrInfinity = require_to_integer_or_infinity();
    var arraySpeciesCreate = require_array_species_create();
    $({ target: "Array", proto: true }, {
      flat: function flat() {
        var depthArg = arguments.length ? arguments[0] : void 0;
        var O = toObject(this);
        var sourceLen = lengthOfArrayLike(O);
        var A = arraySpeciesCreate(O, 0);
        A.length = flattenIntoArray(A, O, O, sourceLen, 0, depthArg === void 0 ? 1 : toIntegerOrInfinity(depthArg));
        return A;
      }
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-keys.js
var require_object_keys = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-keys.js": function(exports2, module2) {
    "use strict";
    var internalObjectKeys = require_object_keys_internal();
    var enumBugKeys = require_enum_bug_keys();
    module2.exports = Object.keys || function keys(O) {
      return internalObjectKeys(O, enumBugKeys);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-define-properties.js
var require_object_define_properties = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-define-properties.js": function(exports2) {
    "use strict";
    var DESCRIPTORS = require_descriptors();
    var V8_PROTOTYPE_DEFINE_BUG = require_v8_prototype_define_bug();
    var definePropertyModule = require_object_define_property();
    var anObject = require_an_object();
    var toIndexedObject = require_to_indexed_object();
    var objectKeys = require_object_keys();
    exports2.f = DESCRIPTORS && !V8_PROTOTYPE_DEFINE_BUG ? Object.defineProperties : function defineProperties(O, Properties) {
      anObject(O);
      var props = toIndexedObject(Properties);
      var keys = objectKeys(Properties);
      var length = keys.length;
      var index = 0;
      var key;
      while (length > index) definePropertyModule.f(O, key = keys[index++], props[key]);
      return O;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/html.js
var require_html = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/html.js": function(exports2, module2) {
    "use strict";
    var getBuiltIn = require_get_built_in();
    module2.exports = getBuiltIn("document", "documentElement");
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-create.js
var require_object_create = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-create.js": function(exports2, module2) {
    "use strict";
    var anObject = require_an_object();
    var definePropertiesModule = require_object_define_properties();
    var enumBugKeys = require_enum_bug_keys();
    var hiddenKeys = require_hidden_keys();
    var html = require_html();
    var documentCreateElement = require_document_create_element();
    var sharedKey = require_shared_key();
    var GT = ">";
    var LT = "<";
    var PROTOTYPE = "prototype";
    var SCRIPT = "script";
    var IE_PROTO = sharedKey("IE_PROTO");
    var EmptyConstructor = function() {
    };
    var scriptTag = function(content) {
      return LT + SCRIPT + GT + content + LT + "/" + SCRIPT + GT;
    };
    var NullProtoObjectViaActiveX = function(activeXDocument2) {
      activeXDocument2.write(scriptTag(""));
      activeXDocument2.close();
      var temp = activeXDocument2.parentWindow.Object;
      activeXDocument2 = null;
      return temp;
    };
    var NullProtoObjectViaIFrame = function() {
      var iframe = documentCreateElement("iframe");
      var JS = "java" + SCRIPT + ":";
      var iframeDocument;
      iframe.style.display = "none";
      html.appendChild(iframe);
      iframe.src = String(JS);
      iframeDocument = iframe.contentWindow.document;
      iframeDocument.open();
      iframeDocument.write(scriptTag("document.F=Object"));
      iframeDocument.close();
      return iframeDocument.F;
    };
    var activeXDocument;
    var NullProtoObject = function() {
      try {
        activeXDocument = new ActiveXObject("htmlfile");
      } catch (error) {
      }
      NullProtoObject = typeof document != "undefined" ? document.domain && activeXDocument ? NullProtoObjectViaActiveX(activeXDocument) : NullProtoObjectViaIFrame() : NullProtoObjectViaActiveX(activeXDocument);
      var length = enumBugKeys.length;
      while (length--) delete NullProtoObject[PROTOTYPE][enumBugKeys[length]];
      return NullProtoObject();
    };
    hiddenKeys[IE_PROTO] = true;
    module2.exports = Object.create || function create(O, Properties) {
      var result;
      if (O !== null) {
        EmptyConstructor[PROTOTYPE] = anObject(O);
        result = new EmptyConstructor();
        EmptyConstructor[PROTOTYPE] = null;
        result[IE_PROTO] = O;
      } else result = NullProtoObject();
      return Properties === void 0 ? result : definePropertiesModule.f(result, Properties);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/add-to-unscopables.js
var require_add_to_unscopables = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/add-to-unscopables.js": function(exports2, module2) {
    "use strict";
    var wellKnownSymbol = require_well_known_symbol();
    var create = require_object_create();
    var defineProperty = require_object_define_property().f;
    var UNSCOPABLES = wellKnownSymbol("unscopables");
    var ArrayPrototype = Array.prototype;
    if (ArrayPrototype[UNSCOPABLES] === void 0) {
      defineProperty(ArrayPrototype, UNSCOPABLES, {
        configurable: true,
        value: create(null)
      });
    }
    module2.exports = function(key) {
      ArrayPrototype[UNSCOPABLES][key] = true;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/iterators.js
var require_iterators = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/iterators.js": function(exports2, module2) {
    "use strict";
    module2.exports = {};
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/correct-prototype-getter.js
var require_correct_prototype_getter = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/correct-prototype-getter.js": function(exports2, module2) {
    "use strict";
    var fails = require_fails();
    module2.exports = !fails(function() {
      function F() {
      }
      F.prototype.constructor = null;
      return Object.getPrototypeOf(new F()) !== F.prototype;
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-get-prototype-of.js
var require_object_get_prototype_of = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-get-prototype-of.js": function(exports2, module2) {
    "use strict";
    var hasOwn = require_has_own_property();
    var isCallable = require_is_callable();
    var toObject = require_to_object();
    var sharedKey = require_shared_key();
    var CORRECT_PROTOTYPE_GETTER = require_correct_prototype_getter();
    var IE_PROTO = sharedKey("IE_PROTO");
    var $Object = Object;
    var ObjectPrototype = $Object.prototype;
    module2.exports = CORRECT_PROTOTYPE_GETTER ? $Object.getPrototypeOf : function(O) {
      var object = toObject(O);
      if (hasOwn(object, IE_PROTO)) return object[IE_PROTO];
      var constructor = object.constructor;
      if (isCallable(constructor) && object instanceof constructor) {
        return constructor.prototype;
      }
      return object instanceof $Object ? ObjectPrototype : null;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/iterators-core.js
var require_iterators_core = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/iterators-core.js": function(exports2, module2) {
    "use strict";
    var fails = require_fails();
    var isCallable = require_is_callable();
    var isObject = require_is_object();
    var create = require_object_create();
    var getPrototypeOf = require_object_get_prototype_of();
    var defineBuiltIn = require_define_built_in();
    var wellKnownSymbol = require_well_known_symbol();
    var IS_PURE = require_is_pure();
    var ITERATOR = wellKnownSymbol("iterator");
    var BUGGY_SAFARI_ITERATORS = false;
    var IteratorPrototype;
    var PrototypeOfArrayIteratorPrototype;
    var arrayIterator;
    if ([].keys) {
      arrayIterator = [].keys();
      if (!("next" in arrayIterator)) BUGGY_SAFARI_ITERATORS = true;
      else {
        PrototypeOfArrayIteratorPrototype = getPrototypeOf(getPrototypeOf(arrayIterator));
        if (PrototypeOfArrayIteratorPrototype !== Object.prototype) IteratorPrototype = PrototypeOfArrayIteratorPrototype;
      }
    }
    var NEW_ITERATOR_PROTOTYPE = !isObject(IteratorPrototype) || fails(function() {
      var test = {};
      return IteratorPrototype[ITERATOR].call(test) !== test;
    });
    if (NEW_ITERATOR_PROTOTYPE) IteratorPrototype = {};
    else if (IS_PURE) IteratorPrototype = create(IteratorPrototype);
    if (!isCallable(IteratorPrototype[ITERATOR])) {
      defineBuiltIn(IteratorPrototype, ITERATOR, function() {
        return this;
      });
    }
    module2.exports = {
      IteratorPrototype: IteratorPrototype,
      BUGGY_SAFARI_ITERATORS: BUGGY_SAFARI_ITERATORS
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/set-to-string-tag.js
var require_set_to_string_tag = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/set-to-string-tag.js": function(exports2, module2) {
    "use strict";
    var defineProperty = require_object_define_property().f;
    var hasOwn = require_has_own_property();
    var wellKnownSymbol = require_well_known_symbol();
    var TO_STRING_TAG = wellKnownSymbol("toStringTag");
    module2.exports = function(target, TAG, STATIC) {
      if (target && !STATIC) target = target.prototype;
      if (target && !hasOwn(target, TO_STRING_TAG)) {
        defineProperty(target, TO_STRING_TAG, { configurable: true, value: TAG });
      }
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/iterator-create-constructor.js
var require_iterator_create_constructor = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/iterator-create-constructor.js": function(exports2, module2) {
    "use strict";
    var IteratorPrototype = require_iterators_core().IteratorPrototype;
    var create = require_object_create();
    var createPropertyDescriptor = require_create_property_descriptor();
    var setToStringTag = require_set_to_string_tag();
    var Iterators = require_iterators();
    var returnThis = function() {
      return this;
    };
    module2.exports = function(IteratorConstructor, NAME, next, ENUMERABLE_NEXT) {
      var TO_STRING_TAG = NAME + " Iterator";
      IteratorConstructor.prototype = create(IteratorPrototype, { next: createPropertyDescriptor(+!ENUMERABLE_NEXT, next) });
      setToStringTag(IteratorConstructor, TO_STRING_TAG, false, true);
      Iterators[TO_STRING_TAG] = returnThis;
      return IteratorConstructor;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/function-uncurry-this-accessor.js
var require_function_uncurry_this_accessor = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/function-uncurry-this-accessor.js": function(exports2, module2) {
    "use strict";
    var uncurryThis = require_function_uncurry_this();
    var aCallable = require_a_callable();
    module2.exports = function(object, key, method) {
      try {
        return uncurryThis(aCallable(Object.getOwnPropertyDescriptor(object, key)[method]));
      } catch (error) {
      }
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-possible-prototype.js
var require_is_possible_prototype = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-possible-prototype.js": function(exports2, module2) {
    "use strict";
    var isObject = require_is_object();
    module2.exports = function(argument) {
      return isObject(argument) || argument === null;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/a-possible-prototype.js
var require_a_possible_prototype = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/a-possible-prototype.js": function(exports2, module2) {
    "use strict";
    var isPossiblePrototype = require_is_possible_prototype();
    var $String = String;
    var $TypeError = TypeError;
    module2.exports = function(argument) {
      if (isPossiblePrototype(argument)) return argument;
      throw new $TypeError("Can't set " + $String(argument) + " as a prototype");
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-set-prototype-of.js
var require_object_set_prototype_of = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-set-prototype-of.js": function(exports2, module2) {
    "use strict";
    var uncurryThisAccessor = require_function_uncurry_this_accessor();
    var isObject = require_is_object();
    var requireObjectCoercible = require_require_object_coercible();
    var aPossiblePrototype = require_a_possible_prototype();
    module2.exports = Object.setPrototypeOf || ("__proto__" in {} ? function() {
      var CORRECT_SETTER = false;
      var test = {};
      var setter;
      try {
        setter = uncurryThisAccessor(Object.prototype, "__proto__", "set");
        setter(test, []);
        CORRECT_SETTER = test instanceof Array;
      } catch (error) {
      }
      return function setPrototypeOf(O, proto) {
        requireObjectCoercible(O);
        aPossiblePrototype(proto);
        if (!isObject(O)) return O;
        if (CORRECT_SETTER) setter(O, proto);
        else O.__proto__ = proto;
        return O;
      };
    }() : void 0);
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/iterator-define.js
var require_iterator_define = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/iterator-define.js": function(exports2, module2) {
    "use strict";
    var $ = require_export();
    var call = require_function_call();
    var IS_PURE = require_is_pure();
    var FunctionName = require_function_name();
    var isCallable = require_is_callable();
    var createIteratorConstructor = require_iterator_create_constructor();
    var getPrototypeOf = require_object_get_prototype_of();
    var setPrototypeOf = require_object_set_prototype_of();
    var setToStringTag = require_set_to_string_tag();
    var createNonEnumerableProperty = require_create_non_enumerable_property();
    var defineBuiltIn = require_define_built_in();
    var wellKnownSymbol = require_well_known_symbol();
    var Iterators = require_iterators();
    var IteratorsCore = require_iterators_core();
    var PROPER_FUNCTION_NAME = FunctionName.PROPER;
    var CONFIGURABLE_FUNCTION_NAME = FunctionName.CONFIGURABLE;
    var IteratorPrototype = IteratorsCore.IteratorPrototype;
    var BUGGY_SAFARI_ITERATORS = IteratorsCore.BUGGY_SAFARI_ITERATORS;
    var ITERATOR = wellKnownSymbol("iterator");
    var KEYS = "keys";
    var VALUES = "values";
    var ENTRIES = "entries";
    var returnThis = function() {
      return this;
    };
    module2.exports = function(Iterable, NAME, IteratorConstructor, next, DEFAULT, IS_SET, FORCED) {
      createIteratorConstructor(IteratorConstructor, NAME, next);
      var getIterationMethod = function(KIND) {
        if (KIND === DEFAULT && defaultIterator) return defaultIterator;
        if (!BUGGY_SAFARI_ITERATORS && KIND && KIND in IterablePrototype) return IterablePrototype[KIND];
        switch (KIND) {
          case KEYS:
            return function keys() {
              return new IteratorConstructor(this, KIND);
            };
          case VALUES:
            return function values() {
              return new IteratorConstructor(this, KIND);
            };
          case ENTRIES:
            return function entries() {
              return new IteratorConstructor(this, KIND);
            };
        }
        return function() {
          return new IteratorConstructor(this);
        };
      };
      var TO_STRING_TAG = NAME + " Iterator";
      var INCORRECT_VALUES_NAME = false;
      var IterablePrototype = Iterable.prototype;
      var nativeIterator = IterablePrototype[ITERATOR] || IterablePrototype["@@iterator"] || DEFAULT && IterablePrototype[DEFAULT];
      var defaultIterator = !BUGGY_SAFARI_ITERATORS && nativeIterator || getIterationMethod(DEFAULT);
      var anyNativeIterator = NAME === "Array" ? IterablePrototype.entries || nativeIterator : nativeIterator;
      var CurrentIteratorPrototype, methods, KEY;
      if (anyNativeIterator) {
        CurrentIteratorPrototype = getPrototypeOf(anyNativeIterator.call(new Iterable()));
        if (CurrentIteratorPrototype !== Object.prototype && CurrentIteratorPrototype.next) {
          if (!IS_PURE && getPrototypeOf(CurrentIteratorPrototype) !== IteratorPrototype) {
            if (setPrototypeOf) {
              setPrototypeOf(CurrentIteratorPrototype, IteratorPrototype);
            } else if (!isCallable(CurrentIteratorPrototype[ITERATOR])) {
              defineBuiltIn(CurrentIteratorPrototype, ITERATOR, returnThis);
            }
          }
          setToStringTag(CurrentIteratorPrototype, TO_STRING_TAG, true, true);
          if (IS_PURE) Iterators[TO_STRING_TAG] = returnThis;
        }
      }
      if (PROPER_FUNCTION_NAME && DEFAULT === VALUES && nativeIterator && nativeIterator.name !== VALUES) {
        if (!IS_PURE && CONFIGURABLE_FUNCTION_NAME) {
          createNonEnumerableProperty(IterablePrototype, "name", VALUES);
        } else {
          INCORRECT_VALUES_NAME = true;
          defaultIterator = function values() {
            return call(nativeIterator, this);
          };
        }
      }
      if (DEFAULT) {
        methods = {
          values: getIterationMethod(VALUES),
          keys: IS_SET ? defaultIterator : getIterationMethod(KEYS),
          entries: getIterationMethod(ENTRIES)
        };
        if (FORCED) for (KEY in methods) {
          if (BUGGY_SAFARI_ITERATORS || INCORRECT_VALUES_NAME || !(KEY in IterablePrototype)) {
            defineBuiltIn(IterablePrototype, KEY, methods[KEY]);
          }
        }
        else $({ target: NAME, proto: true, forced: BUGGY_SAFARI_ITERATORS || INCORRECT_VALUES_NAME }, methods);
      }
      if ((!IS_PURE || FORCED) && IterablePrototype[ITERATOR] !== defaultIterator) {
        defineBuiltIn(IterablePrototype, ITERATOR, defaultIterator, { name: DEFAULT });
      }
      Iterators[NAME] = defaultIterator;
      return methods;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/create-iter-result-object.js
var require_create_iter_result_object = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/create-iter-result-object.js": function(exports2, module2) {
    "use strict";
    module2.exports = function(value, done) {
      return { value: value, done: done };
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.array.iterator.js
var require_es_array_iterator = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.array.iterator.js": function(exports2, module2) {
    "use strict";
    var toIndexedObject = require_to_indexed_object();
    var addToUnscopables = require_add_to_unscopables();
    var Iterators = require_iterators();
    var InternalStateModule = require_internal_state();
    var defineProperty = require_object_define_property().f;
    var defineIterator = require_iterator_define();
    var createIterResultObject = require_create_iter_result_object();
    var IS_PURE = require_is_pure();
    var DESCRIPTORS = require_descriptors();
    var ARRAY_ITERATOR = "Array Iterator";
    var setInternalState = InternalStateModule.set;
    var getInternalState = InternalStateModule.getterFor(ARRAY_ITERATOR);
    module2.exports = defineIterator(Array, "Array", function(iterated, kind) {
      setInternalState(this, {
        type: ARRAY_ITERATOR,
        target: toIndexedObject(iterated),
        // target
        index: 0,
        // next index
        kind: kind
        // kind
      });
    }, function() {
      var state = getInternalState(this);
      var target = state.target;
      var index = state.index++;
      if (!target || index >= target.length) {
        state.target = null;
        return createIterResultObject(void 0, true);
      }
      switch (state.kind) {
        case "keys":
          return createIterResultObject(index, false);
        case "values":
          return createIterResultObject(target[index], false);
      }
      return createIterResultObject([index, target[index]], false);
    }, "values");
    var values = Iterators.Arguments = Iterators.Array;
    addToUnscopables("keys");
    addToUnscopables("values");
    addToUnscopables("entries");
    if (!IS_PURE && DESCRIPTORS && values.name !== "values") try {
      defineProperty(values, "name", { value: "values" });
    } catch (error) {
    }
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.array.unscopables.flat.js
var require_es_array_unscopables_flat = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.array.unscopables.flat.js": function() {
    "use strict";
    var addToUnscopables = require_add_to_unscopables();
    addToUnscopables("flat");
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-array-iterator-method.js
var require_is_array_iterator_method = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-array-iterator-method.js": function(exports2, module2) {
    "use strict";
    var wellKnownSymbol = require_well_known_symbol();
    var Iterators = require_iterators();
    var ITERATOR = wellKnownSymbol("iterator");
    var ArrayPrototype = Array.prototype;
    module2.exports = function(it) {
      return it !== void 0 && (Iterators.Array === it || ArrayPrototype[ITERATOR] === it);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/get-iterator-method.js
var require_get_iterator_method = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/get-iterator-method.js": function(exports2, module2) {
    "use strict";
    var classof = require_classof();
    var getMethod = require_get_method();
    var isNullOrUndefined = require_is_null_or_undefined();
    var Iterators = require_iterators();
    var wellKnownSymbol = require_well_known_symbol();
    var ITERATOR = wellKnownSymbol("iterator");
    module2.exports = function(it) {
      if (!isNullOrUndefined(it)) return getMethod(it, ITERATOR) || getMethod(it, "@@iterator") || Iterators[classof(it)];
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/get-iterator.js
var require_get_iterator = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/get-iterator.js": function(exports2, module2) {
    "use strict";
    var call = require_function_call();
    var aCallable = require_a_callable();
    var anObject = require_an_object();
    var tryToString = require_try_to_string();
    var getIteratorMethod = require_get_iterator_method();
    var $TypeError = TypeError;
    module2.exports = function(argument, usingIterator) {
      var iteratorMethod = arguments.length < 2 ? getIteratorMethod(argument) : usingIterator;
      if (aCallable(iteratorMethod)) return anObject(call(iteratorMethod, argument));
      throw new $TypeError(tryToString(argument) + " is not iterable");
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/iterator-close.js
var require_iterator_close = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/iterator-close.js": function(exports2, module2) {
    "use strict";
    var call = require_function_call();
    var anObject = require_an_object();
    var getMethod = require_get_method();
    module2.exports = function(iterator, kind, value) {
      var innerResult, innerError;
      anObject(iterator);
      try {
        innerResult = getMethod(iterator, "return");
        if (!innerResult) {
          if (kind === "throw") throw value;
          return value;
        }
        innerResult = call(innerResult, iterator);
      } catch (error) {
        innerError = true;
        innerResult = error;
      }
      if (kind === "throw") throw value;
      if (innerError) throw innerResult;
      anObject(innerResult);
      return value;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/iterate.js
var require_iterate = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/iterate.js": function(exports2, module2) {
    "use strict";
    var bind = require_function_bind_context();
    var call = require_function_call();
    var anObject = require_an_object();
    var tryToString = require_try_to_string();
    var isArrayIteratorMethod = require_is_array_iterator_method();
    var lengthOfArrayLike = require_length_of_array_like();
    var isPrototypeOf = require_object_is_prototype_of();
    var getIterator = require_get_iterator();
    var getIteratorMethod = require_get_iterator_method();
    var iteratorClose = require_iterator_close();
    var $TypeError = TypeError;
    var Result = function(stopped, result) {
      this.stopped = stopped;
      this.result = result;
    };
    var ResultPrototype = Result.prototype;
    module2.exports = function(iterable, unboundFunction, options) {
      var that = options && options.that;
      var AS_ENTRIES = !!(options && options.AS_ENTRIES);
      var IS_RECORD = !!(options && options.IS_RECORD);
      var IS_ITERATOR = !!(options && options.IS_ITERATOR);
      var INTERRUPTED = !!(options && options.INTERRUPTED);
      var fn = bind(unboundFunction, that);
      var iterator, iterFn, index, length, result, next, step;
      var stop = function(condition) {
        if (iterator) iteratorClose(iterator, "normal", condition);
        return new Result(true, condition);
      };
      var callFn = function(value) {
        if (AS_ENTRIES) {
          anObject(value);
          return INTERRUPTED ? fn(value[0], value[1], stop) : fn(value[0], value[1]);
        }
        return INTERRUPTED ? fn(value, stop) : fn(value);
      };
      if (IS_RECORD) {
        iterator = iterable.iterator;
      } else if (IS_ITERATOR) {
        iterator = iterable;
      } else {
        iterFn = getIteratorMethod(iterable);
        if (!iterFn) throw new $TypeError(tryToString(iterable) + " is not iterable");
        if (isArrayIteratorMethod(iterFn)) {
          for (index = 0, length = lengthOfArrayLike(iterable); length > index; index++) {
            result = callFn(iterable[index]);
            if (result && isPrototypeOf(ResultPrototype, result)) return result;
          }
          return new Result(false);
        }
        iterator = getIterator(iterable, iterFn);
      }
      next = IS_RECORD ? iterable.next : iterator.next;
      while (!(step = call(next, iterator)).done) {
        try {
          result = callFn(step.value);
        } catch (error) {
          iteratorClose(iterator, "throw", error);
        }
        if (typeof result == "object" && result && isPrototypeOf(ResultPrototype, result)) return result;
      }
      return new Result(false);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/create-property.js
var require_create_property = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/create-property.js": function(exports2, module2) {
    "use strict";
    var DESCRIPTORS = require_descriptors();
    var definePropertyModule = require_object_define_property();
    var createPropertyDescriptor = require_create_property_descriptor();
    module2.exports = function(object, key, value) {
      if (DESCRIPTORS) definePropertyModule.f(object, key, createPropertyDescriptor(0, value));
      else object[key] = value;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.object.from-entries.js
var require_es_object_from_entries = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.object.from-entries.js": function() {
    "use strict";
    var $ = require_export();
    var iterate = require_iterate();
    var createProperty = require_create_property();
    $({ target: "Object", stat: true }, {
      fromEntries: function fromEntries(iterable) {
        var obj = {};
        iterate(iterable, function(k, v) {
          createProperty(obj, k, v);
        }, { AS_ENTRIES: true });
        return obj;
      }
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/environment.js
var require_environment = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/environment.js": function(exports2, module2) {
    "use strict";
    var globalThis2 = require_global_this();
    var userAgent = require_environment_user_agent();
    var classof = require_classof_raw();
    var userAgentStartsWith = function(string) {
      return userAgent.slice(0, string.length) === string;
    };
    module2.exports = function() {
      if (userAgentStartsWith("Bun/")) return "BUN";
      if (userAgentStartsWith("Cloudflare-Workers")) return "CLOUDFLARE";
      if (userAgentStartsWith("Deno/")) return "DENO";
      if (userAgentStartsWith("Node.js/")) return "NODE";
      if (globalThis2.Bun && typeof Bun.version == "string") return "BUN";
      if (globalThis2.Deno && typeof Deno.version == "object") return "DENO";
      if (classof(globalThis2.process) === "process") return "NODE";
      if (globalThis2.window && globalThis2.document) return "BROWSER";
      return "REST";
    }();
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/environment-is-node.js
var require_environment_is_node = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/environment-is-node.js": function(exports2, module2) {
    "use strict";
    var ENVIRONMENT = require_environment();
    module2.exports = ENVIRONMENT === "NODE";
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/set-species.js
var require_set_species = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/set-species.js": function(exports2, module2) {
    "use strict";
    var getBuiltIn = require_get_built_in();
    var defineBuiltInAccessor = require_define_built_in_accessor();
    var wellKnownSymbol = require_well_known_symbol();
    var DESCRIPTORS = require_descriptors();
    var SPECIES = wellKnownSymbol("species");
    module2.exports = function(CONSTRUCTOR_NAME) {
      var Constructor = getBuiltIn(CONSTRUCTOR_NAME);
      if (DESCRIPTORS && Constructor && !Constructor[SPECIES]) {
        defineBuiltInAccessor(Constructor, SPECIES, {
          configurable: true,
          get: function() {
            return this;
          }
        });
      }
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/an-instance.js
var require_an_instance = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/an-instance.js": function(exports2, module2) {
    "use strict";
    var isPrototypeOf = require_object_is_prototype_of();
    var $TypeError = TypeError;
    module2.exports = function(it, Prototype) {
      if (isPrototypeOf(Prototype, it)) return it;
      throw new $TypeError("Incorrect invocation");
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/a-constructor.js
var require_a_constructor = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/a-constructor.js": function(exports2, module2) {
    "use strict";
    var isConstructor = require_is_constructor();
    var tryToString = require_try_to_string();
    var $TypeError = TypeError;
    module2.exports = function(argument) {
      if (isConstructor(argument)) return argument;
      throw new $TypeError(tryToString(argument) + " is not a constructor");
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/species-constructor.js
var require_species_constructor = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/species-constructor.js": function(exports2, module2) {
    "use strict";
    var anObject = require_an_object();
    var aConstructor = require_a_constructor();
    var isNullOrUndefined = require_is_null_or_undefined();
    var wellKnownSymbol = require_well_known_symbol();
    var SPECIES = wellKnownSymbol("species");
    module2.exports = function(O, defaultConstructor) {
      var C = anObject(O).constructor;
      var S;
      return C === void 0 || isNullOrUndefined(S = anObject(C)[SPECIES]) ? defaultConstructor : aConstructor(S);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/function-apply.js
var require_function_apply = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/function-apply.js": function(exports2, module2) {
    "use strict";
    var NATIVE_BIND = require_function_bind_native();
    var FunctionPrototype = Function.prototype;
    var apply = FunctionPrototype.apply;
    var call = FunctionPrototype.call;
    module2.exports = typeof Reflect == "object" && Reflect.apply || (NATIVE_BIND ? call.bind(apply) : function() {
      return call.apply(apply, arguments);
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/array-slice.js
var require_array_slice = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/array-slice.js": function(exports2, module2) {
    "use strict";
    var uncurryThis = require_function_uncurry_this();
    module2.exports = uncurryThis([].slice);
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/validate-arguments-length.js
var require_validate_arguments_length = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/validate-arguments-length.js": function(exports2, module2) {
    "use strict";
    var $TypeError = TypeError;
    module2.exports = function(passed, required) {
      if (passed < required) throw new $TypeError("Not enough arguments");
      return passed;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/environment-is-ios.js
var require_environment_is_ios = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/environment-is-ios.js": function(exports2, module2) {
    "use strict";
    var userAgent = require_environment_user_agent();
    module2.exports = /(?:ipad|iphone|ipod).*applewebkit/i.test(userAgent);
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/task.js
var require_task = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/task.js": function(exports2, module2) {
    "use strict";
    var globalThis2 = require_global_this();
    var apply = require_function_apply();
    var bind = require_function_bind_context();
    var isCallable = require_is_callable();
    var hasOwn = require_has_own_property();
    var fails = require_fails();
    var html = require_html();
    var arraySlice = require_array_slice();
    var createElement = require_document_create_element();
    var validateArgumentsLength = require_validate_arguments_length();
    var IS_IOS = require_environment_is_ios();
    var IS_NODE = require_environment_is_node();
    var set = globalThis2.setImmediate;
    var clear = globalThis2.clearImmediate;
    var process = globalThis2.process;
    var Dispatch = globalThis2.Dispatch;
    var Function2 = globalThis2.Function;
    var MessageChannel = globalThis2.MessageChannel;
    var String2 = globalThis2.String;
    var counter = 0;
    var queue = {};
    var ONREADYSTATECHANGE = "onreadystatechange";
    var $location;
    var defer;
    var channel;
    var port;
    fails(function() {
      $location = globalThis2.location;
    });
    var run = function(id) {
      if (hasOwn(queue, id)) {
        var fn = queue[id];
        delete queue[id];
        fn();
      }
    };
    var runner = function(id) {
      return function() {
        run(id);
      };
    };
    var eventListener = function(event2) {
      run(event2.data);
    };
    var globalPostMessageDefer = function(id) {
      globalThis2.postMessage(String2(id), $location.protocol + "//" + $location.host);
    };
    if (!set || !clear) {
      set = function setImmediate(handler) {
        validateArgumentsLength(arguments.length, 1);
        var fn = isCallable(handler) ? handler : Function2(handler);
        var args = arraySlice(arguments, 1);
        queue[++counter] = function() {
          apply(fn, void 0, args);
        };
        defer(counter);
        return counter;
      };
      clear = function clearImmediate(id) {
        delete queue[id];
      };
      if (IS_NODE) {
        defer = function(id) {
          process.nextTick(runner(id));
        };
      } else if (Dispatch && Dispatch.now) {
        defer = function(id) {
          Dispatch.now(runner(id));
        };
      } else if (MessageChannel && !IS_IOS) {
        channel = new MessageChannel();
        port = channel.port2;
        channel.port1.onmessage = eventListener;
        defer = bind(port.postMessage, port);
      } else if (globalThis2.addEventListener && isCallable(globalThis2.postMessage) && !globalThis2.importScripts && $location && $location.protocol !== "file:" && !fails(globalPostMessageDefer)) {
        defer = globalPostMessageDefer;
        globalThis2.addEventListener("message", eventListener, false);
      } else if (ONREADYSTATECHANGE in createElement("script")) {
        defer = function(id) {
          html.appendChild(createElement("script"))[ONREADYSTATECHANGE] = function() {
            html.removeChild(this);
            run(id);
          };
        };
      } else {
        defer = function(id) {
          setTimeout(runner(id), 0);
        };
      }
    }
    module2.exports = {
      set: set,
      clear: clear
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/safe-get-built-in.js
var require_safe_get_built_in = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/safe-get-built-in.js": function(exports2, module2) {
    "use strict";
    var globalThis2 = require_global_this();
    var DESCRIPTORS = require_descriptors();
    var getOwnPropertyDescriptor = Object.getOwnPropertyDescriptor;
    module2.exports = function(name) {
      if (!DESCRIPTORS) return globalThis2[name];
      var descriptor = getOwnPropertyDescriptor(globalThis2, name);
      return descriptor && descriptor.value;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/queue.js
var require_queue = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/queue.js": function(exports2, module2) {
    "use strict";
    var Queue = function() {
      this.head = null;
      this.tail = null;
    };
    Queue.prototype = {
      add: function(item) {
        var entry = { item: item, next: null };
        var tail = this.tail;
        if (tail) tail.next = entry;
        else this.head = entry;
        this.tail = entry;
      },
      get: function() {
        var entry = this.head;
        if (entry) {
          var next = this.head = entry.next;
          if (next === null) this.tail = null;
          return entry.item;
        }
      }
    };
    module2.exports = Queue;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/environment-is-ios-pebble.js
var require_environment_is_ios_pebble = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/environment-is-ios-pebble.js": function(exports2, module2) {
    "use strict";
    var userAgent = require_environment_user_agent();
    module2.exports = /ipad|iphone|ipod/i.test(userAgent) && typeof Pebble != "undefined";
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/environment-is-webos-webkit.js
var require_environment_is_webos_webkit = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/environment-is-webos-webkit.js": function(exports2, module2) {
    "use strict";
    var userAgent = require_environment_user_agent();
    module2.exports = /web0s(?!.*chrome)/i.test(userAgent);
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/microtask.js
var require_microtask = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/microtask.js": function(exports2, module2) {
    "use strict";
    var globalThis2 = require_global_this();
    var safeGetBuiltIn = require_safe_get_built_in();
    var bind = require_function_bind_context();
    var macrotask = require_task().set;
    var Queue = require_queue();
    var IS_IOS = require_environment_is_ios();
    var IS_IOS_PEBBLE = require_environment_is_ios_pebble();
    var IS_WEBOS_WEBKIT = require_environment_is_webos_webkit();
    var IS_NODE = require_environment_is_node();
    var MutationObserver = globalThis2.MutationObserver || globalThis2.WebKitMutationObserver;
    var document2 = globalThis2.document;
    var process = globalThis2.process;
    var Promise2 = globalThis2.Promise;
    var microtask = safeGetBuiltIn("queueMicrotask");
    var notify;
    var toggle;
    var node;
    var promise;
    var then;
    if (!microtask) {
      queue = new Queue();
      flush = function() {
        var parent, fn;
        if (IS_NODE && (parent = process.domain)) parent.exit();
        while (fn = queue.get()) try {
          fn();
        } catch (error) {
          if (queue.head) notify();
          throw error;
        }
        if (parent) parent.enter();
      };
      if (!IS_IOS && !IS_NODE && !IS_WEBOS_WEBKIT && MutationObserver && document2) {
        toggle = true;
        node = document2.createTextNode("");
        new MutationObserver(flush).observe(node, { characterData: true });
        notify = function() {
          node.data = toggle = !toggle;
        };
      } else if (!IS_IOS_PEBBLE && Promise2 && Promise2.resolve) {
        promise = Promise2.resolve(void 0);
        promise.constructor = Promise2;
        then = bind(promise.then, promise);
        notify = function() {
          then(flush);
        };
      } else if (IS_NODE) {
        notify = function() {
          process.nextTick(flush);
        };
      } else {
        macrotask = bind(macrotask, globalThis2);
        notify = function() {
          macrotask(flush);
        };
      }
      microtask = function(fn) {
        if (!queue.head) notify();
        queue.add(fn);
      };
    }
    var queue;
    var flush;
    module2.exports = microtask;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/host-report-errors.js
var require_host_report_errors = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/host-report-errors.js": function(exports2, module2) {
    "use strict";
    module2.exports = function(a, b) {
      try {
        arguments.length === 1 ? console.error(a) : console.error(a, b);
      } catch (error) {
      }
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/perform.js
var require_perform = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/perform.js": function(exports2, module2) {
    "use strict";
    module2.exports = function(exec) {
      try {
        return { error: false, value: exec() };
      } catch (error) {
        return { error: true, value: error };
      }
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/promise-native-constructor.js
var require_promise_native_constructor = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/promise-native-constructor.js": function(exports2, module2) {
    "use strict";
    var globalThis2 = require_global_this();
    module2.exports = globalThis2.Promise;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/promise-constructor-detection.js
var require_promise_constructor_detection = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/promise-constructor-detection.js": function(exports2, module2) {
    "use strict";
    var globalThis2 = require_global_this();
    var NativePromiseConstructor = require_promise_native_constructor();
    var isCallable = require_is_callable();
    var isForced = require_is_forced();
    var inspectSource = require_inspect_source();
    var wellKnownSymbol = require_well_known_symbol();
    var ENVIRONMENT = require_environment();
    var IS_PURE = require_is_pure();
    var V8_VERSION = require_environment_v8_version();
    var NativePromisePrototype = NativePromiseConstructor && NativePromiseConstructor.prototype;
    var SPECIES = wellKnownSymbol("species");
    var SUBCLASSING = false;
    var NATIVE_PROMISE_REJECTION_EVENT = isCallable(globalThis2.PromiseRejectionEvent);
    var FORCED_PROMISE_CONSTRUCTOR = isForced("Promise", function() {
      var PROMISE_CONSTRUCTOR_SOURCE = inspectSource(NativePromiseConstructor);
      var GLOBAL_CORE_JS_PROMISE = PROMISE_CONSTRUCTOR_SOURCE !== String(NativePromiseConstructor);
      if (!GLOBAL_CORE_JS_PROMISE && V8_VERSION === 66) return true;
      if (IS_PURE && !(NativePromisePrototype["catch"] && NativePromisePrototype["finally"])) return true;
      if (!V8_VERSION || V8_VERSION < 51 || !/native code/.test(PROMISE_CONSTRUCTOR_SOURCE)) {
        var promise = new NativePromiseConstructor(function(resolve) {
          resolve(1);
        });
        var FakePromise = function(exec) {
          exec(function() {
          }, function() {
          });
        };
        var constructor = promise.constructor = {};
        constructor[SPECIES] = FakePromise;
        SUBCLASSING = promise.then(function() {
        }) instanceof FakePromise;
        if (!SUBCLASSING) return true;
      }
      return !GLOBAL_CORE_JS_PROMISE && (ENVIRONMENT === "BROWSER" || ENVIRONMENT === "DENO") && !NATIVE_PROMISE_REJECTION_EVENT;
    });
    module2.exports = {
      CONSTRUCTOR: FORCED_PROMISE_CONSTRUCTOR,
      REJECTION_EVENT: NATIVE_PROMISE_REJECTION_EVENT,
      SUBCLASSING: SUBCLASSING
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/new-promise-capability.js
var require_new_promise_capability = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/new-promise-capability.js": function(exports2, module2) {
    "use strict";
    var aCallable = require_a_callable();
    var $TypeError = TypeError;
    var PromiseCapability = function(C) {
      var resolve, reject;
      this.promise = new C(function($$resolve, $$reject) {
        if (resolve !== void 0 || reject !== void 0) throw new $TypeError("Bad Promise constructor");
        resolve = $$resolve;
        reject = $$reject;
      });
      this.resolve = aCallable(resolve);
      this.reject = aCallable(reject);
    };
    module2.exports.f = function(C) {
      return new PromiseCapability(C);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.promise.constructor.js
var require_es_promise_constructor = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.promise.constructor.js": function() {
    "use strict";
    var $ = require_export();
    var IS_PURE = require_is_pure();
    var IS_NODE = require_environment_is_node();
    var globalThis2 = require_global_this();
    var call = require_function_call();
    var defineBuiltIn = require_define_built_in();
    var setPrototypeOf = require_object_set_prototype_of();
    var setToStringTag = require_set_to_string_tag();
    var setSpecies = require_set_species();
    var aCallable = require_a_callable();
    var isCallable = require_is_callable();
    var isObject = require_is_object();
    var anInstance = require_an_instance();
    var speciesConstructor = require_species_constructor();
    var task = require_task().set;
    var microtask = require_microtask();
    var hostReportErrors = require_host_report_errors();
    var perform = require_perform();
    var Queue = require_queue();
    var InternalStateModule = require_internal_state();
    var NativePromiseConstructor = require_promise_native_constructor();
    var PromiseConstructorDetection = require_promise_constructor_detection();
    var newPromiseCapabilityModule = require_new_promise_capability();
    var PROMISE = "Promise";
    var FORCED_PROMISE_CONSTRUCTOR = PromiseConstructorDetection.CONSTRUCTOR;
    var NATIVE_PROMISE_REJECTION_EVENT = PromiseConstructorDetection.REJECTION_EVENT;
    var NATIVE_PROMISE_SUBCLASSING = PromiseConstructorDetection.SUBCLASSING;
    var getInternalPromiseState = InternalStateModule.getterFor(PROMISE);
    var setInternalState = InternalStateModule.set;
    var NativePromisePrototype = NativePromiseConstructor && NativePromiseConstructor.prototype;
    var PromiseConstructor = NativePromiseConstructor;
    var PromisePrototype = NativePromisePrototype;
    var TypeError2 = globalThis2.TypeError;
    var document2 = globalThis2.document;
    var process = globalThis2.process;
    var newPromiseCapability = newPromiseCapabilityModule.f;
    var newGenericPromiseCapability = newPromiseCapability;
    var DISPATCH_EVENT = !!(document2 && document2.createEvent && globalThis2.dispatchEvent);
    var UNHANDLED_REJECTION = "unhandledrejection";
    var REJECTION_HANDLED = "rejectionhandled";
    var PENDING = 0;
    var FULFILLED = 1;
    var REJECTED = 2;
    var HANDLED = 1;
    var UNHANDLED = 2;
    var Internal;
    var OwnPromiseCapability;
    var PromiseWrapper;
    var nativeThen;
    var isThenable = function(it) {
      var then;
      return isObject(it) && isCallable(then = it.then) ? then : false;
    };
    var callReaction = function(reaction, state) {
      var value = state.value;
      var ok = state.state === FULFILLED;
      var handler = ok ? reaction.ok : reaction.fail;
      var resolve = reaction.resolve;
      var reject = reaction.reject;
      var domain = reaction.domain;
      var result, then, exited;
      try {
        if (handler) {
          if (!ok) {
            if (state.rejection === UNHANDLED) onHandleUnhandled(state);
            state.rejection = HANDLED;
          }
          if (handler === true) result = value;
          else {
            if (domain) domain.enter();
            result = handler(value);
            if (domain) {
              domain.exit();
              exited = true;
            }
          }
          if (result === reaction.promise) {
            reject(new TypeError2("Promise-chain cycle"));
          } else if (then = isThenable(result)) {
            call(then, result, resolve, reject);
          } else resolve(result);
        } else reject(value);
      } catch (error) {
        if (domain && !exited) domain.exit();
        reject(error);
      }
    };
    var notify = function(state, isReject) {
      if (state.notified) return;
      state.notified = true;
      microtask(function() {
        var reactions = state.reactions;
        var reaction;
        while (reaction = reactions.get()) {
          callReaction(reaction, state);
        }
        state.notified = false;
        if (isReject && !state.rejection) onUnhandled(state);
      });
    };
    var dispatchEvent = function(name, promise, reason) {
      var event2, handler;
      if (DISPATCH_EVENT) {
        event2 = document2.createEvent("Event");
        event2.promise = promise;
        event2.reason = reason;
        event2.initEvent(name, false, true);
        globalThis2.dispatchEvent(event2);
      } else event2 = { promise: promise, reason: reason };
      if (!NATIVE_PROMISE_REJECTION_EVENT && (handler = globalThis2["on" + name])) handler(event2);
      else if (name === UNHANDLED_REJECTION) hostReportErrors("Unhandled promise rejection", reason);
    };
    var onUnhandled = function(state) {
      call(task, globalThis2, function() {
        var promise = state.facade;
        var value = state.value;
        var IS_UNHANDLED = isUnhandled(state);
        var result;
        if (IS_UNHANDLED) {
          result = perform(function() {
            if (IS_NODE) {
              process.emit("unhandledRejection", value, promise);
            } else dispatchEvent(UNHANDLED_REJECTION, promise, value);
          });
          state.rejection = IS_NODE || isUnhandled(state) ? UNHANDLED : HANDLED;
          if (result.error) throw result.value;
        }
      });
    };
    var isUnhandled = function(state) {
      return state.rejection !== HANDLED && !state.parent;
    };
    var onHandleUnhandled = function(state) {
      call(task, globalThis2, function() {
        var promise = state.facade;
        if (IS_NODE) {
          process.emit("rejectionHandled", promise);
        } else dispatchEvent(REJECTION_HANDLED, promise, state.value);
      });
    };
    var bind = function(fn, state, unwrap) {
      return function(value) {
        fn(state, value, unwrap);
      };
    };
    var internalReject = function(state, value, unwrap) {
      if (state.done) return;
      state.done = true;
      if (unwrap) state = unwrap;
      state.value = value;
      state.state = REJECTED;
      notify(state, true);
    };
    var internalResolve = function(state, value, unwrap) {
      if (state.done) return;
      state.done = true;
      if (unwrap) state = unwrap;
      try {
        if (state.facade === value) throw new TypeError2("Promise can't be resolved itself");
        var then = isThenable(value);
        if (then) {
          microtask(function() {
            var wrapper = { done: false };
            try {
              call(
                then,
                value,
                bind(internalResolve, wrapper, state),
                bind(internalReject, wrapper, state)
              );
            } catch (error) {
              internalReject(wrapper, error, state);
            }
          });
        } else {
          state.value = value;
          state.state = FULFILLED;
          notify(state, false);
        }
      } catch (error) {
        internalReject({ done: false }, error, state);
      }
    };
    if (FORCED_PROMISE_CONSTRUCTOR) {
      PromiseConstructor = function Promise2(executor) {
        anInstance(this, PromisePrototype);
        aCallable(executor);
        call(Internal, this);
        var state = getInternalPromiseState(this);
        try {
          executor(bind(internalResolve, state), bind(internalReject, state));
        } catch (error) {
          internalReject(state, error);
        }
      };
      PromisePrototype = PromiseConstructor.prototype;
      Internal = function Promise2(executor) {
        setInternalState(this, {
          type: PROMISE,
          done: false,
          notified: false,
          parent: false,
          reactions: new Queue(),
          rejection: false,
          state: PENDING,
          value: null
        });
      };
      Internal.prototype = defineBuiltIn(PromisePrototype, "then", function then(onFulfilled, onRejected) {
        var state = getInternalPromiseState(this);
        var reaction = newPromiseCapability(speciesConstructor(this, PromiseConstructor));
        state.parent = true;
        reaction.ok = isCallable(onFulfilled) ? onFulfilled : true;
        reaction.fail = isCallable(onRejected) && onRejected;
        reaction.domain = IS_NODE ? process.domain : void 0;
        if (state.state === PENDING) state.reactions.add(reaction);
        else microtask(function() {
          callReaction(reaction, state);
        });
        return reaction.promise;
      });
      OwnPromiseCapability = function() {
        var promise = new Internal();
        var state = getInternalPromiseState(promise);
        this.promise = promise;
        this.resolve = bind(internalResolve, state);
        this.reject = bind(internalReject, state);
      };
      newPromiseCapabilityModule.f = newPromiseCapability = function(C) {
        return C === PromiseConstructor || C === PromiseWrapper ? new OwnPromiseCapability(C) : newGenericPromiseCapability(C);
      };
      if (!IS_PURE && isCallable(NativePromiseConstructor) && NativePromisePrototype !== Object.prototype) {
        nativeThen = NativePromisePrototype.then;
        if (!NATIVE_PROMISE_SUBCLASSING) {
          defineBuiltIn(NativePromisePrototype, "then", function then(onFulfilled, onRejected) {
            var that = this;
            return new PromiseConstructor(function(resolve, reject) {
              call(nativeThen, that, resolve, reject);
            }).then(onFulfilled, onRejected);
          }, { unsafe: true });
        }
        try {
          delete NativePromisePrototype.constructor;
        } catch (error) {
        }
        if (setPrototypeOf) {
          setPrototypeOf(NativePromisePrototype, PromisePrototype);
        }
      }
    }
    $({ global: true, constructor: true, wrap: true, forced: FORCED_PROMISE_CONSTRUCTOR }, {
      Promise: PromiseConstructor
    });
    setToStringTag(PromiseConstructor, PROMISE, false, true);
    setSpecies(PROMISE);
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/check-correctness-of-iteration.js
var require_check_correctness_of_iteration = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/check-correctness-of-iteration.js": function(exports2, module2) {
    "use strict";
    var wellKnownSymbol = require_well_known_symbol();
    var ITERATOR = wellKnownSymbol("iterator");
    var SAFE_CLOSING = false;
    try {
      called = 0;
      iteratorWithReturn = {
        next: function() {
          return { done: !!called++ };
        },
        "return": function() {
          SAFE_CLOSING = true;
        }
      };
      iteratorWithReturn[ITERATOR] = function() {
        return this;
      };
      Array.from(iteratorWithReturn, function() {
        throw 2;
      });
    } catch (error) {
    }
    var called;
    var iteratorWithReturn;
    module2.exports = function(exec, SKIP_CLOSING) {
      try {
        if (!SKIP_CLOSING && !SAFE_CLOSING) return false;
      } catch (error) {
        return false;
      }
      var ITERATION_SUPPORT = false;
      try {
        var object = {};
        object[ITERATOR] = function() {
          return {
            next: function() {
              return { done: ITERATION_SUPPORT = true };
            }
          };
        };
        exec(object);
      } catch (error) {
      }
      return ITERATION_SUPPORT;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/promise-statics-incorrect-iteration.js
var require_promise_statics_incorrect_iteration = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/promise-statics-incorrect-iteration.js": function(exports2, module2) {
    "use strict";
    var NativePromiseConstructor = require_promise_native_constructor();
    var checkCorrectnessOfIteration = require_check_correctness_of_iteration();
    var FORCED_PROMISE_CONSTRUCTOR = require_promise_constructor_detection().CONSTRUCTOR;
    module2.exports = FORCED_PROMISE_CONSTRUCTOR || !checkCorrectnessOfIteration(function(iterable) {
      NativePromiseConstructor.all(iterable).then(void 0, function() {
      });
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.promise.all.js
var require_es_promise_all = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.promise.all.js": function() {
    "use strict";
    var $ = require_export();
    var call = require_function_call();
    var aCallable = require_a_callable();
    var newPromiseCapabilityModule = require_new_promise_capability();
    var perform = require_perform();
    var iterate = require_iterate();
    var PROMISE_STATICS_INCORRECT_ITERATION = require_promise_statics_incorrect_iteration();
    $({ target: "Promise", stat: true, forced: PROMISE_STATICS_INCORRECT_ITERATION }, {
      all: function all(iterable) {
        var C = this;
        var capability = newPromiseCapabilityModule.f(C);
        var resolve = capability.resolve;
        var reject = capability.reject;
        var result = perform(function() {
          var $promiseResolve = aCallable(C.resolve);
          var values = [];
          var counter = 0;
          var remaining = 1;
          iterate(iterable, function(promise) {
            var index = counter++;
            var alreadyCalled = false;
            remaining++;
            call($promiseResolve, C, promise).then(function(value) {
              if (alreadyCalled) return;
              alreadyCalled = true;
              values[index] = value;
              --remaining || resolve(values);
            }, reject);
          });
          --remaining || resolve(values);
        });
        if (result.error) reject(result.value);
        return capability.promise;
      }
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.promise.catch.js
var require_es_promise_catch = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.promise.catch.js": function() {
    "use strict";
    var $ = require_export();
    var IS_PURE = require_is_pure();
    var FORCED_PROMISE_CONSTRUCTOR = require_promise_constructor_detection().CONSTRUCTOR;
    var NativePromiseConstructor = require_promise_native_constructor();
    var getBuiltIn = require_get_built_in();
    var isCallable = require_is_callable();
    var defineBuiltIn = require_define_built_in();
    var NativePromisePrototype = NativePromiseConstructor && NativePromiseConstructor.prototype;
    $({ target: "Promise", proto: true, forced: FORCED_PROMISE_CONSTRUCTOR, real: true }, {
      "catch": function(onRejected) {
        return this.then(void 0, onRejected);
      }
    });
    if (!IS_PURE && isCallable(NativePromiseConstructor)) {
      method = getBuiltIn("Promise").prototype["catch"];
      if (NativePromisePrototype["catch"] !== method) {
        defineBuiltIn(NativePromisePrototype, "catch", method, { unsafe: true });
      }
    }
    var method;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.promise.race.js
var require_es_promise_race = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.promise.race.js": function() {
    "use strict";
    var $ = require_export();
    var call = require_function_call();
    var aCallable = require_a_callable();
    var newPromiseCapabilityModule = require_new_promise_capability();
    var perform = require_perform();
    var iterate = require_iterate();
    var PROMISE_STATICS_INCORRECT_ITERATION = require_promise_statics_incorrect_iteration();
    $({ target: "Promise", stat: true, forced: PROMISE_STATICS_INCORRECT_ITERATION }, {
      race: function race(iterable) {
        var C = this;
        var capability = newPromiseCapabilityModule.f(C);
        var reject = capability.reject;
        var result = perform(function() {
          var $promiseResolve = aCallable(C.resolve);
          iterate(iterable, function(promise) {
            call($promiseResolve, C, promise).then(capability.resolve, reject);
          });
        });
        if (result.error) reject(result.value);
        return capability.promise;
      }
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.promise.reject.js
var require_es_promise_reject = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.promise.reject.js": function() {
    "use strict";
    var $ = require_export();
    var newPromiseCapabilityModule = require_new_promise_capability();
    var FORCED_PROMISE_CONSTRUCTOR = require_promise_constructor_detection().CONSTRUCTOR;
    $({ target: "Promise", stat: true, forced: FORCED_PROMISE_CONSTRUCTOR }, {
      reject: function reject(r) {
        var capability = newPromiseCapabilityModule.f(this);
        var capabilityReject = capability.reject;
        capabilityReject(r);
        return capability.promise;
      }
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/promise-resolve.js
var require_promise_resolve = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/promise-resolve.js": function(exports2, module2) {
    "use strict";
    var anObject = require_an_object();
    var isObject = require_is_object();
    var newPromiseCapability = require_new_promise_capability();
    module2.exports = function(C, x) {
      anObject(C);
      if (isObject(x) && x.constructor === C) return x;
      var promiseCapability = newPromiseCapability.f(C);
      var resolve = promiseCapability.resolve;
      resolve(x);
      return promiseCapability.promise;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.promise.resolve.js
var require_es_promise_resolve = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.promise.resolve.js": function() {
    "use strict";
    var $ = require_export();
    var getBuiltIn = require_get_built_in();
    var IS_PURE = require_is_pure();
    var NativePromiseConstructor = require_promise_native_constructor();
    var FORCED_PROMISE_CONSTRUCTOR = require_promise_constructor_detection().CONSTRUCTOR;
    var promiseResolve = require_promise_resolve();
    var PromiseConstructorWrapper = getBuiltIn("Promise");
    var CHECK_WRAPPER = IS_PURE && !FORCED_PROMISE_CONSTRUCTOR;
    $({ target: "Promise", stat: true, forced: IS_PURE || FORCED_PROMISE_CONSTRUCTOR }, {
      resolve: function resolve(x) {
        return promiseResolve(CHECK_WRAPPER && this === PromiseConstructorWrapper ? NativePromiseConstructor : this, x);
      }
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.promise.js
var require_es_promise = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.promise.js": function() {
    "use strict";
    require_es_promise_constructor();
    require_es_promise_all();
    require_es_promise_catch();
    require_es_promise_race();
    require_es_promise_reject();
    require_es_promise_resolve();
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/inherit-if-required.js
var require_inherit_if_required = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/inherit-if-required.js": function(exports2, module2) {
    "use strict";
    var isCallable = require_is_callable();
    var isObject = require_is_object();
    var setPrototypeOf = require_object_set_prototype_of();
    module2.exports = function($this, dummy, Wrapper) {
      var NewTarget, NewTargetPrototype;
      if (
        // it can work only with native `setPrototypeOf`
        setPrototypeOf && // we haven't completely correct pre-ES6 way for getting `new.target`, so use this
        isCallable(NewTarget = dummy.constructor) && NewTarget !== Wrapper && isObject(NewTargetPrototype = NewTarget.prototype) && NewTargetPrototype !== Wrapper.prototype
      ) setPrototypeOf($this, NewTargetPrototype);
      return $this;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-regexp.js
var require_is_regexp = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/is-regexp.js": function(exports2, module2) {
    "use strict";
    var isObject = require_is_object();
    var classof = require_classof_raw();
    var wellKnownSymbol = require_well_known_symbol();
    var MATCH = wellKnownSymbol("match");
    module2.exports = function(it) {
      var isRegExp;
      return isObject(it) && ((isRegExp = it[MATCH]) !== void 0 ? !!isRegExp : classof(it) === "RegExp");
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/regexp-flags.js
var require_regexp_flags = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/regexp-flags.js": function(exports2, module2) {
    "use strict";
    var anObject = require_an_object();
    module2.exports = function() {
      var that = anObject(this);
      var result = "";
      if (that.hasIndices) result += "d";
      if (that.global) result += "g";
      if (that.ignoreCase) result += "i";
      if (that.multiline) result += "m";
      if (that.dotAll) result += "s";
      if (that.unicode) result += "u";
      if (that.unicodeSets) result += "v";
      if (that.sticky) result += "y";
      return result;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/regexp-get-flags.js
var require_regexp_get_flags = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/regexp-get-flags.js": function(exports2, module2) {
    "use strict";
    var call = require_function_call();
    var hasOwn = require_has_own_property();
    var isPrototypeOf = require_object_is_prototype_of();
    var regExpFlags = require_regexp_flags();
    var RegExpPrototype = RegExp.prototype;
    module2.exports = function(R) {
      var flags = R.flags;
      return flags === void 0 && !("flags" in RegExpPrototype) && !hasOwn(R, "flags") && isPrototypeOf(RegExpPrototype, R) ? call(regExpFlags, R) : flags;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/regexp-sticky-helpers.js
var require_regexp_sticky_helpers = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/regexp-sticky-helpers.js": function(exports2, module2) {
    "use strict";
    var fails = require_fails();
    var globalThis2 = require_global_this();
    var $RegExp = globalThis2.RegExp;
    var UNSUPPORTED_Y = fails(function() {
      var re = $RegExp("a", "y");
      re.lastIndex = 2;
      return re.exec("abcd") !== null;
    });
    var MISSED_STICKY = UNSUPPORTED_Y || fails(function() {
      return !$RegExp("a", "y").sticky;
    });
    var BROKEN_CARET = UNSUPPORTED_Y || fails(function() {
      var re = $RegExp("^r", "gy");
      re.lastIndex = 2;
      return re.exec("str") !== null;
    });
    module2.exports = {
      BROKEN_CARET: BROKEN_CARET,
      MISSED_STICKY: MISSED_STICKY,
      UNSUPPORTED_Y: UNSUPPORTED_Y
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/proxy-accessor.js
var require_proxy_accessor = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/proxy-accessor.js": function(exports2, module2) {
    "use strict";
    var defineProperty = require_object_define_property().f;
    module2.exports = function(Target, Source, key) {
      key in Target || defineProperty(Target, key, {
        configurable: true,
        get: function() {
          return Source[key];
        },
        set: function(it) {
          Source[key] = it;
        }
      });
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/regexp-unsupported-dot-all.js
var require_regexp_unsupported_dot_all = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/regexp-unsupported-dot-all.js": function(exports2, module2) {
    "use strict";
    var fails = require_fails();
    var globalThis2 = require_global_this();
    var $RegExp = globalThis2.RegExp;
    module2.exports = fails(function() {
      var re = $RegExp(".", "s");
      return !(re.dotAll && re.test("\n") && re.flags === "s");
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/regexp-unsupported-ncg.js
var require_regexp_unsupported_ncg = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/regexp-unsupported-ncg.js": function(exports2, module2) {
    "use strict";
    var fails = require_fails();
    var globalThis2 = require_global_this();
    var $RegExp = globalThis2.RegExp;
    module2.exports = fails(function() {
      var re = $RegExp("(?<a>b)", "g");
      return re.exec("b").groups.a !== "b" || "b".replace(re, "$<a>c") !== "bc";
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.regexp.constructor.js
var require_es_regexp_constructor = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.regexp.constructor.js": function() {
    "use strict";
    var DESCRIPTORS = require_descriptors();
    var globalThis2 = require_global_this();
    var uncurryThis = require_function_uncurry_this();
    var isForced = require_is_forced();
    var inheritIfRequired = require_inherit_if_required();
    var createNonEnumerableProperty = require_create_non_enumerable_property();
    var create = require_object_create();
    var getOwnPropertyNames = require_object_get_own_property_names().f;
    var isPrototypeOf = require_object_is_prototype_of();
    var isRegExp = require_is_regexp();
    var toString = require_to_string();
    var getRegExpFlags = require_regexp_get_flags();
    var stickyHelpers = require_regexp_sticky_helpers();
    var proxyAccessor = require_proxy_accessor();
    var defineBuiltIn = require_define_built_in();
    var fails = require_fails();
    var hasOwn = require_has_own_property();
    var enforceInternalState = require_internal_state().enforce;
    var setSpecies = require_set_species();
    var wellKnownSymbol = require_well_known_symbol();
    var UNSUPPORTED_DOT_ALL = require_regexp_unsupported_dot_all();
    var UNSUPPORTED_NCG = require_regexp_unsupported_ncg();
    var MATCH = wellKnownSymbol("match");
    var NativeRegExp = globalThis2.RegExp;
    var RegExpPrototype = NativeRegExp.prototype;
    var SyntaxError = globalThis2.SyntaxError;
    var exec = uncurryThis(RegExpPrototype.exec);
    var charAt = uncurryThis("".charAt);
    var replace = uncurryThis("".replace);
    var stringIndexOf = uncurryThis("".indexOf);
    var stringSlice = uncurryThis("".slice);
    var IS_NCG = /^\?<[^\s\d!#%&*+<=>@^][^\s!#%&*+<=>@^]*>/;
    var re1 = /a/g;
    var re2 = /a/g;
    var CORRECT_NEW = new NativeRegExp(re1) !== re1;
    var MISSED_STICKY = stickyHelpers.MISSED_STICKY;
    var UNSUPPORTED_Y = stickyHelpers.UNSUPPORTED_Y;
    var BASE_FORCED = DESCRIPTORS && (!CORRECT_NEW || MISSED_STICKY || UNSUPPORTED_DOT_ALL || UNSUPPORTED_NCG || fails(function() {
      re2[MATCH] = false;
      return NativeRegExp(re1) !== re1 || NativeRegExp(re2) === re2 || String(NativeRegExp(re1, "i")) !== "/a/i";
    }));
    var handleDotAll = function(string) {
      var length = string.length;
      var index2 = 0;
      var result = "";
      var brackets = false;
      var chr;
      for (; index2 <= length; index2++) {
        chr = charAt(string, index2);
        if (chr === "\\") {
          result += chr + charAt(string, ++index2);
          continue;
        }
        if (!brackets && chr === ".") {
          result += "[\\s\\S]";
        } else {
          if (chr === "[") {
            brackets = true;
          } else if (chr === "]") {
            brackets = false;
          }
          result += chr;
        }
      }
      return result;
    };
    var handleNCG = function(string) {
      var length = string.length;
      var index2 = 0;
      var result = "";
      var named = [];
      var names = create(null);
      var brackets = false;
      var ncg = false;
      var groupid = 0;
      var groupname = "";
      var chr;
      for (; index2 <= length; index2++) {
        chr = charAt(string, index2);
        if (chr === "\\") {
          chr += charAt(string, ++index2);
        } else if (chr === "]") {
          brackets = false;
        } else if (!brackets) switch (true) {
          case chr === "[":
            brackets = true;
            break;
          case chr === "(":
            result += chr;
            if (stringSlice(string, index2 + 1, index2 + 3) === "?:") {
              continue;
            }
            if (exec(IS_NCG, stringSlice(string, index2 + 1))) {
              index2 += 2;
              ncg = true;
            }
            groupid++;
            continue;
          case (chr === ">" && ncg):
            if (groupname === "" || hasOwn(names, groupname)) {
              throw new SyntaxError("Invalid capture group name");
            }
            names[groupname] = true;
            named[named.length] = [groupname, groupid];
            ncg = false;
            groupname = "";
            continue;
        }
        if (ncg) groupname += chr;
        else result += chr;
      }
      return [result, named];
    };
    if (isForced("RegExp", BASE_FORCED)) {
      RegExpWrapper = function RegExp2(pattern, flags) {
        var thisIsRegExp = isPrototypeOf(RegExpPrototype, this);
        var patternIsRegExp = isRegExp(pattern);
        var flagsAreUndefined = flags === void 0;
        var groups = [];
        var rawPattern = pattern;
        var rawFlags, dotAll, sticky, handled, result, state;
        if (!thisIsRegExp && patternIsRegExp && flagsAreUndefined && pattern.constructor === RegExpWrapper) {
          return pattern;
        }
        if (patternIsRegExp || isPrototypeOf(RegExpPrototype, pattern)) {
          pattern = pattern.source;
          if (flagsAreUndefined) flags = getRegExpFlags(rawPattern);
        }
        pattern = pattern === void 0 ? "" : toString(pattern);
        flags = flags === void 0 ? "" : toString(flags);
        rawPattern = pattern;
        if (UNSUPPORTED_DOT_ALL && "dotAll" in re1) {
          dotAll = !!flags && stringIndexOf(flags, "s") > -1;
          if (dotAll) flags = replace(flags, /s/g, "");
        }
        rawFlags = flags;
        if (MISSED_STICKY && "sticky" in re1) {
          sticky = !!flags && stringIndexOf(flags, "y") > -1;
          if (sticky && UNSUPPORTED_Y) flags = replace(flags, /y/g, "");
        }
        if (UNSUPPORTED_NCG) {
          handled = handleNCG(pattern);
          pattern = handled[0];
          groups = handled[1];
        }
        result = inheritIfRequired(NativeRegExp(pattern, flags), thisIsRegExp ? this : RegExpPrototype, RegExpWrapper);
        if (dotAll || sticky || groups.length) {
          state = enforceInternalState(result);
          if (dotAll) {
            state.dotAll = true;
            state.raw = RegExpWrapper(handleDotAll(pattern), rawFlags);
          }
          if (sticky) state.sticky = true;
          if (groups.length) state.groups = groups;
        }
        if (pattern !== rawPattern) try {
          createNonEnumerableProperty(result, "source", rawPattern === "" ? "(?:)" : rawPattern);
        } catch (error) {
        }
        return result;
      };
      for (keys = getOwnPropertyNames(NativeRegExp), index = 0; keys.length > index; ) {
        proxyAccessor(RegExpWrapper, NativeRegExp, keys[index++]);
      }
      RegExpPrototype.constructor = RegExpWrapper;
      RegExpWrapper.prototype = RegExpPrototype;
      defineBuiltIn(globalThis2, "RegExp", RegExpWrapper, { constructor: true });
    }
    var RegExpWrapper;
    var keys;
    var index;
    setSpecies("RegExp");
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/regexp-exec.js
var require_regexp_exec = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/regexp-exec.js": function(exports2, module2) {
    "use strict";
    var call = require_function_call();
    var uncurryThis = require_function_uncurry_this();
    var toString = require_to_string();
    var regexpFlags = require_regexp_flags();
    var stickyHelpers = require_regexp_sticky_helpers();
    var shared = require_shared();
    var create = require_object_create();
    var getInternalState = require_internal_state().get;
    var UNSUPPORTED_DOT_ALL = require_regexp_unsupported_dot_all();
    var UNSUPPORTED_NCG = require_regexp_unsupported_ncg();
    var nativeReplace = shared("native-string-replace", String.prototype.replace);
    var nativeExec = RegExp.prototype.exec;
    var patchedExec = nativeExec;
    var charAt = uncurryThis("".charAt);
    var indexOf = uncurryThis("".indexOf);
    var replace = uncurryThis("".replace);
    var stringSlice = uncurryThis("".slice);
    var UPDATES_LAST_INDEX_WRONG = function() {
      var re1 = /a/;
      var re2 = /b*/g;
      call(nativeExec, re1, "a");
      call(nativeExec, re2, "a");
      return re1.lastIndex !== 0 || re2.lastIndex !== 0;
    }();
    var UNSUPPORTED_Y = stickyHelpers.BROKEN_CARET;
    var NPCG_INCLUDED = /()??/.exec("")[1] !== void 0;
    var PATCH = UPDATES_LAST_INDEX_WRONG || NPCG_INCLUDED || UNSUPPORTED_Y || UNSUPPORTED_DOT_ALL || UNSUPPORTED_NCG;
    if (PATCH) {
      patchedExec = function exec(string) {
        var re = this;
        var state = getInternalState(re);
        var str2 = toString(string);
        var raw = state.raw;
        var result, reCopy, lastIndex, match, i, object, group;
        if (raw) {
          raw.lastIndex = re.lastIndex;
          result = call(patchedExec, raw, str2);
          re.lastIndex = raw.lastIndex;
          return result;
        }
        var groups = state.groups;
        var sticky = UNSUPPORTED_Y && re.sticky;
        var flags = call(regexpFlags, re);
        var source = re.source;
        var charsAdded = 0;
        var strCopy = str2;
        if (sticky) {
          flags = replace(flags, "y", "");
          if (indexOf(flags, "g") === -1) {
            flags += "g";
          }
          strCopy = stringSlice(str2, re.lastIndex);
          if (re.lastIndex > 0 && (!re.multiline || re.multiline && charAt(str2, re.lastIndex - 1) !== "\n")) {
            source = "(?: " + source + ")";
            strCopy = " " + strCopy;
            charsAdded++;
          }
          reCopy = new RegExp("^(?:" + source + ")", flags);
        }
        if (NPCG_INCLUDED) {
          reCopy = new RegExp("^" + source + "$(?!\\s)", flags);
        }
        if (UPDATES_LAST_INDEX_WRONG) lastIndex = re.lastIndex;
        match = call(nativeExec, sticky ? reCopy : re, strCopy);
        if (sticky) {
          if (match) {
            match.input = stringSlice(match.input, charsAdded);
            match[0] = stringSlice(match[0], charsAdded);
            match.index = re.lastIndex;
            re.lastIndex += match[0].length;
          } else re.lastIndex = 0;
        } else if (UPDATES_LAST_INDEX_WRONG && match) {
          re.lastIndex = re.global ? match.index + match[0].length : lastIndex;
        }
        if (NPCG_INCLUDED && match && match.length > 1) {
          call(nativeReplace, match[0], reCopy, function() {
            for (i = 1; i < arguments.length - 2; i++) {
              if (arguments[i] === void 0) match[i] = void 0;
            }
          });
        }
        if (match && groups) {
          match.groups = object = create(null);
          for (i = 0; i < groups.length; i++) {
            group = groups[i];
            object[group[0]] = match[group[1]];
          }
        }
        return match;
      };
    }
    module2.exports = patchedExec;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.regexp.exec.js
var require_es_regexp_exec = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.regexp.exec.js": function() {
    "use strict";
    var $ = require_export();
    var exec = require_regexp_exec();
    $({ target: "RegExp", proto: true, forced: /./.exec !== exec }, {
      exec: exec
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/fix-regexp-well-known-symbol-logic.js
var require_fix_regexp_well_known_symbol_logic = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/fix-regexp-well-known-symbol-logic.js": function(exports2, module2) {
    "use strict";
    require_es_regexp_exec();
    var call = require_function_call();
    var defineBuiltIn = require_define_built_in();
    var regexpExec = require_regexp_exec();
    var fails = require_fails();
    var wellKnownSymbol = require_well_known_symbol();
    var createNonEnumerableProperty = require_create_non_enumerable_property();
    var SPECIES = wellKnownSymbol("species");
    var RegExpPrototype = RegExp.prototype;
    module2.exports = function(KEY, exec, FORCED, SHAM) {
      var SYMBOL = wellKnownSymbol(KEY);
      var DELEGATES_TO_SYMBOL = !fails(function() {
        var O = {};
        O[SYMBOL] = function() {
          return 7;
        };
        return ""[KEY](O) !== 7;
      });
      var DELEGATES_TO_EXEC = DELEGATES_TO_SYMBOL && !fails(function() {
        var execCalled = false;
        var re = /a/;
        if (KEY === "split") {
          re = {};
          re.constructor = {};
          re.constructor[SPECIES] = function() {
            return re;
          };
          re.flags = "";
          re[SYMBOL] = /./[SYMBOL];
        }
        re.exec = function() {
          execCalled = true;
          return null;
        };
        re[SYMBOL]("");
        return !execCalled;
      });
      if (!DELEGATES_TO_SYMBOL || !DELEGATES_TO_EXEC || FORCED) {
        var nativeRegExpMethod = /./[SYMBOL];
        var methods = exec(SYMBOL, ""[KEY], function(nativeMethod, regexp, str2, arg2, forceStringMethod) {
          var $exec = regexp.exec;
          if ($exec === regexpExec || $exec === RegExpPrototype.exec) {
            if (DELEGATES_TO_SYMBOL && !forceStringMethod) {
              return { done: true, value: call(nativeRegExpMethod, regexp, str2, arg2) };
            }
            return { done: true, value: call(nativeMethod, str2, regexp, arg2) };
          }
          return { done: false };
        });
        defineBuiltIn(String.prototype, KEY, methods[0]);
        defineBuiltIn(RegExpPrototype, SYMBOL, methods[1]);
      }
      if (SHAM) createNonEnumerableProperty(RegExpPrototype[SYMBOL], "sham", true);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/string-multibyte.js
var require_string_multibyte = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/string-multibyte.js": function(exports2, module2) {
    "use strict";
    var uncurryThis = require_function_uncurry_this();
    var toIntegerOrInfinity = require_to_integer_or_infinity();
    var toString = require_to_string();
    var requireObjectCoercible = require_require_object_coercible();
    var charAt = uncurryThis("".charAt);
    var charCodeAt = uncurryThis("".charCodeAt);
    var stringSlice = uncurryThis("".slice);
    var createMethod = function(CONVERT_TO_STRING) {
      return function($this, pos) {
        var S = toString(requireObjectCoercible($this));
        var position = toIntegerOrInfinity(pos);
        var size = S.length;
        var first, second;
        if (position < 0 || position >= size) return CONVERT_TO_STRING ? "" : void 0;
        first = charCodeAt(S, position);
        return first < 55296 || first > 56319 || position + 1 === size || (second = charCodeAt(S, position + 1)) < 56320 || second > 57343 ? CONVERT_TO_STRING ? charAt(S, position) : first : CONVERT_TO_STRING ? stringSlice(S, position, position + 2) : (first - 55296 << 10) + (second - 56320) + 65536;
      };
    };
    module2.exports = {
      // `String.prototype.codePointAt` method
      // https://tc39.es/ecma262/#sec-string.prototype.codepointat
      codeAt: createMethod(false),
      // `String.prototype.at` method
      // https://github.com/mathiasbynens/String.prototype.at
      charAt: createMethod(true)
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/advance-string-index.js
var require_advance_string_index = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/advance-string-index.js": function(exports2, module2) {
    "use strict";
    var charAt = require_string_multibyte().charAt;
    module2.exports = function(S, index, unicode) {
      return index + (unicode ? charAt(S, index).length : 1);
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/get-substitution.js
var require_get_substitution = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/get-substitution.js": function(exports2, module2) {
    "use strict";
    var uncurryThis = require_function_uncurry_this();
    var toObject = require_to_object();
    var floor = Math.floor;
    var charAt = uncurryThis("".charAt);
    var replace = uncurryThis("".replace);
    var stringSlice = uncurryThis("".slice);
    var SUBSTITUTION_SYMBOLS = /\$([$&'`]|\d{1,2}|<[^>]*>)/g;
    var SUBSTITUTION_SYMBOLS_NO_NAMED = /\$([$&'`]|\d{1,2})/g;
    module2.exports = function(matched, str2, position, captures, namedCaptures, replacement) {
      var tailPos = position + matched.length;
      var m = captures.length;
      var symbols = SUBSTITUTION_SYMBOLS_NO_NAMED;
      if (namedCaptures !== void 0) {
        namedCaptures = toObject(namedCaptures);
        symbols = SUBSTITUTION_SYMBOLS;
      }
      return replace(replacement, symbols, function(match, ch) {
        var capture;
        switch (charAt(ch, 0)) {
          case "$":
            return "$";
          case "&":
            return matched;
          case "`":
            return stringSlice(str2, 0, position);
          case "'":
            return stringSlice(str2, tailPos);
          case "<":
            capture = namedCaptures[stringSlice(ch, 1, -1)];
            break;
          default:
            var n = +ch;
            if (n === 0) return match;
            if (n > m) {
              var f = floor(n / 10);
              if (f === 0) return match;
              if (f <= m) return captures[f - 1] === void 0 ? charAt(ch, 1) : captures[f - 1] + charAt(ch, 1);
              return match;
            }
            capture = captures[n - 1];
        }
        return capture === void 0 ? "" : capture;
      });
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/regexp-exec-abstract.js
var require_regexp_exec_abstract = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/regexp-exec-abstract.js": function(exports2, module2) {
    "use strict";
    var call = require_function_call();
    var anObject = require_an_object();
    var isCallable = require_is_callable();
    var classof = require_classof_raw();
    var regexpExec = require_regexp_exec();
    var $TypeError = TypeError;
    module2.exports = function(R, S) {
      var exec = R.exec;
      if (isCallable(exec)) {
        var result = call(exec, R, S);
        if (result !== null) anObject(result);
        return result;
      }
      if (classof(R) === "RegExp") return call(regexpExec, R, S);
      throw new $TypeError("RegExp#exec called on incompatible receiver");
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.string.replace.js
var require_es_string_replace = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.string.replace.js": function() {
    "use strict";
    var apply = require_function_apply();
    var call = require_function_call();
    var uncurryThis = require_function_uncurry_this();
    var fixRegExpWellKnownSymbolLogic = require_fix_regexp_well_known_symbol_logic();
    var fails = require_fails();
    var anObject = require_an_object();
    var isCallable = require_is_callable();
    var isNullOrUndefined = require_is_null_or_undefined();
    var toIntegerOrInfinity = require_to_integer_or_infinity();
    var toLength = require_to_length();
    var toString = require_to_string();
    var requireObjectCoercible = require_require_object_coercible();
    var advanceStringIndex = require_advance_string_index();
    var getMethod = require_get_method();
    var getSubstitution = require_get_substitution();
    var regExpExec = require_regexp_exec_abstract();
    var wellKnownSymbol = require_well_known_symbol();
    var REPLACE = wellKnownSymbol("replace");
    var max = Math.max;
    var min = Math.min;
    var concat = uncurryThis([].concat);
    var push = uncurryThis([].push);
    var stringIndexOf = uncurryThis("".indexOf);
    var stringSlice = uncurryThis("".slice);
    var maybeToString = function(it) {
      return it === void 0 ? it : String(it);
    };
    var REPLACE_KEEPS_$0 = function() {
      return "a".replace(/./, "$0") === "$0";
    }();
    var REGEXP_REPLACE_SUBSTITUTES_UNDEFINED_CAPTURE = function() {
      if (/./[REPLACE]) {
        return /./[REPLACE]("a", "$0") === "";
      }
      return false;
    }();
    var REPLACE_SUPPORTS_NAMED_GROUPS = !fails(function() {
      var re = /./;
      re.exec = function() {
        var result = [];
        result.groups = { a: "7" };
        return result;
      };
      return "".replace(re, "$<a>") !== "7";
    });
    fixRegExpWellKnownSymbolLogic("replace", function(_, nativeReplace, maybeCallNative) {
      var UNSAFE_SUBSTITUTE = REGEXP_REPLACE_SUBSTITUTES_UNDEFINED_CAPTURE ? "$" : "$0";
      return [
        // `String.prototype.replace` method
        // https://tc39.es/ecma262/#sec-string.prototype.replace
        function replace(searchValue, replaceValue) {
          var O = requireObjectCoercible(this);
          var replacer = isNullOrUndefined(searchValue) ? void 0 : getMethod(searchValue, REPLACE);
          return replacer ? call(replacer, searchValue, O, replaceValue) : call(nativeReplace, toString(O), searchValue, replaceValue);
        },
        // `RegExp.prototype[@@replace]` method
        // https://tc39.es/ecma262/#sec-regexp.prototype-@@replace
        function(string, replaceValue) {
          var rx = anObject(this);
          var S = toString(string);
          if (typeof replaceValue == "string" && stringIndexOf(replaceValue, UNSAFE_SUBSTITUTE) === -1 && stringIndexOf(replaceValue, "$<") === -1) {
            var res = maybeCallNative(nativeReplace, rx, S, replaceValue);
            if (res.done) return res.value;
          }
          var functionalReplace = isCallable(replaceValue);
          if (!functionalReplace) replaceValue = toString(replaceValue);
          var global2 = rx.global;
          var fullUnicode;
          if (global2) {
            fullUnicode = rx.unicode;
            rx.lastIndex = 0;
          }
          var results = [];
          var result;
          while (true) {
            result = regExpExec(rx, S);
            if (result === null) break;
            push(results, result);
            if (!global2) break;
            var matchStr = toString(result[0]);
            if (matchStr === "") rx.lastIndex = advanceStringIndex(S, toLength(rx.lastIndex), fullUnicode);
          }
          var accumulatedResult = "";
          var nextSourcePosition = 0;
          for (var i = 0; i < results.length; i++) {
            result = results[i];
            var matched = toString(result[0]);
            var position = max(min(toIntegerOrInfinity(result.index), S.length), 0);
            var captures = [];
            var replacement;
            for (var j = 1; j < result.length; j++) push(captures, maybeToString(result[j]));
            var namedCaptures = result.groups;
            if (functionalReplace) {
              var replacerArgs = concat([matched], captures, position, S);
              if (namedCaptures !== void 0) push(replacerArgs, namedCaptures);
              replacement = toString(apply(replaceValue, void 0, replacerArgs));
            } else {
              replacement = getSubstitution(matched, S, position, captures, namedCaptures, replaceValue);
            }
            if (position >= nextSourcePosition) {
              accumulatedResult += stringSlice(S, nextSourcePosition, position) + replacement;
              nextSourcePosition = position + matched.length;
            }
          }
          return accumulatedResult + stringSlice(S, nextSourcePosition);
        }
      ];
    }, !REPLACE_SUPPORTS_NAMED_GROUPS || !REPLACE_KEEPS_$0 || REGEXP_REPLACE_SUBSTITUTES_UNDEFINED_CAPTURE);
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.string.split.js
var require_es_string_split = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.string.split.js": function() {
    "use strict";
    var call = require_function_call();
    var uncurryThis = require_function_uncurry_this();
    var fixRegExpWellKnownSymbolLogic = require_fix_regexp_well_known_symbol_logic();
    var anObject = require_an_object();
    var isNullOrUndefined = require_is_null_or_undefined();
    var requireObjectCoercible = require_require_object_coercible();
    var speciesConstructor = require_species_constructor();
    var advanceStringIndex = require_advance_string_index();
    var toLength = require_to_length();
    var toString = require_to_string();
    var getMethod = require_get_method();
    var regExpExec = require_regexp_exec_abstract();
    var stickyHelpers = require_regexp_sticky_helpers();
    var fails = require_fails();
    var UNSUPPORTED_Y = stickyHelpers.UNSUPPORTED_Y;
    var MAX_UINT32 = 4294967295;
    var min = Math.min;
    var push = uncurryThis([].push);
    var stringSlice = uncurryThis("".slice);
    var SPLIT_WORKS_WITH_OVERWRITTEN_EXEC = !fails(function() {
      var re = /(?:)/;
      var originalExec = re.exec;
      re.exec = function() {
        return originalExec.apply(this, arguments);
      };
      var result = "ab".split(re);
      return result.length !== 2 || result[0] !== "a" || result[1] !== "b";
    });
    var BUGGY = "abbc".split(/(b)*/)[1] === "c" || // eslint-disable-next-line regexp/no-empty-group -- required for testing
    "test".split(/(?:)/, -1).length !== 4 || "ab".split(/(?:ab)*/).length !== 2 || ".".split(/(.?)(.?)/).length !== 4 || // eslint-disable-next-line regexp/no-empty-capturing-group, regexp/no-empty-group -- required for testing
    ".".split(/()()/).length > 1 || "".split(/.?/).length;
    fixRegExpWellKnownSymbolLogic("split", function(SPLIT, nativeSplit, maybeCallNative) {
      var internalSplit = "0".split(void 0, 0).length ? function(separator, limit) {
        return separator === void 0 && limit === 0 ? [] : call(nativeSplit, this, separator, limit);
      } : nativeSplit;
      return [
        // `String.prototype.split` method
        // https://tc39.es/ecma262/#sec-string.prototype.split
        function split(separator, limit) {
          var O = requireObjectCoercible(this);
          var splitter = isNullOrUndefined(separator) ? void 0 : getMethod(separator, SPLIT);
          return splitter ? call(splitter, separator, O, limit) : call(internalSplit, toString(O), separator, limit);
        },
        // `RegExp.prototype[@@split]` method
        // https://tc39.es/ecma262/#sec-regexp.prototype-@@split
        //
        // NOTE: This cannot be properly polyfilled in engines that don't support
        // the 'y' flag.
        function(string, limit) {
          var rx = anObject(this);
          var S = toString(string);
          if (!BUGGY) {
            var res = maybeCallNative(internalSplit, rx, S, limit, internalSplit !== nativeSplit);
            if (res.done) return res.value;
          }
          var C = speciesConstructor(rx, RegExp);
          var unicodeMatching = rx.unicode;
          var flags = (rx.ignoreCase ? "i" : "") + (rx.multiline ? "m" : "") + (rx.unicode ? "u" : "") + (UNSUPPORTED_Y ? "g" : "y");
          var splitter = new C(UNSUPPORTED_Y ? "^(?:" + rx.source + ")" : rx, flags);
          var lim = limit === void 0 ? MAX_UINT32 : limit >>> 0;
          if (lim === 0) return [];
          if (S.length === 0) return regExpExec(splitter, S) === null ? [S] : [];
          var p = 0;
          var q = 0;
          var A = [];
          while (q < S.length) {
            splitter.lastIndex = UNSUPPORTED_Y ? 0 : q;
            var z = regExpExec(splitter, UNSUPPORTED_Y ? stringSlice(S, q) : S);
            var e;
            if (z === null || (e = min(toLength(splitter.lastIndex + (UNSUPPORTED_Y ? q : 0)), S.length)) === p) {
              q = advanceStringIndex(S, q, unicodeMatching);
            } else {
              push(A, stringSlice(S, p, q));
              if (A.length === lim) return A;
              for (var i = 1; i <= z.length - 1; i++) {
                push(A, z[i]);
                if (A.length === lim) return A;
              }
              q = p = e;
            }
          }
          push(A, stringSlice(S, p));
          return A;
        }
      ];
    }, BUGGY || !SPLIT_WORKS_WITH_OVERWRITTEN_EXEC, UNSUPPORTED_Y);
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/whitespaces.js
var require_whitespaces = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/whitespaces.js": function(exports2, module2) {
    "use strict";
    module2.exports = "	\n\v\f\r \xA0\u1680\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200A\u202F\u205F\u3000\u2028\u2029\uFEFF";
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/string-trim.js
var require_string_trim = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/string-trim.js": function(exports2, module2) {
    "use strict";
    var uncurryThis = require_function_uncurry_this();
    var requireObjectCoercible = require_require_object_coercible();
    var toString = require_to_string();
    var whitespaces = require_whitespaces();
    var replace = uncurryThis("".replace);
    var ltrim = RegExp("^[" + whitespaces + "]+");
    var rtrim = RegExp("(^|[^" + whitespaces + "])[" + whitespaces + "]+$");
    var createMethod = function(TYPE) {
      return function($this) {
        var string = toString(requireObjectCoercible($this));
        if (TYPE & 1) string = replace(string, ltrim, "");
        if (TYPE & 2) string = replace(string, rtrim, "$1");
        return string;
      };
    };
    module2.exports = {
      // `String.prototype.{ trimLeft, trimStart }` methods
      // https://tc39.es/ecma262/#sec-string.prototype.trimstart
      start: createMethod(1),
      // `String.prototype.{ trimRight, trimEnd }` methods
      // https://tc39.es/ecma262/#sec-string.prototype.trimend
      end: createMethod(2),
      // `String.prototype.trim` method
      // https://tc39.es/ecma262/#sec-string.prototype.trim
      trim: createMethod(3)
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/string-trim-forced.js
var require_string_trim_forced = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/string-trim-forced.js": function(exports2, module2) {
    "use strict";
    var PROPER_FUNCTION_NAME = require_function_name().PROPER;
    var fails = require_fails();
    var whitespaces = require_whitespaces();
    var non = "\u200B\x85\u180E";
    module2.exports = function(METHOD_NAME) {
      return fails(function() {
        return !!whitespaces[METHOD_NAME]() || non[METHOD_NAME]() !== non || PROPER_FUNCTION_NAME && whitespaces[METHOD_NAME].name !== METHOD_NAME;
      });
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.string.trim.js
var require_es_string_trim = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.string.trim.js": function() {
    "use strict";
    var $ = require_export();
    var $trim = require_string_trim().trim;
    var forcedStringTrimMethod = require_string_trim_forced();
    $({ target: "String", proto: true, forced: forcedStringTrimMethod("trim") }, {
      trim: function trim() {
        return $trim(this);
      }
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/dom-iterables.js
var require_dom_iterables = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/dom-iterables.js": function(exports2, module2) {
    "use strict";
    module2.exports = {
      CSSRuleList: 0,
      CSSStyleDeclaration: 0,
      CSSValueList: 0,
      ClientRectList: 0,
      DOMRectList: 0,
      DOMStringList: 0,
      DOMTokenList: 1,
      DataTransferItemList: 0,
      FileList: 0,
      HTMLAllCollection: 0,
      HTMLCollection: 0,
      HTMLFormElement: 0,
      HTMLSelectElement: 0,
      MediaList: 0,
      MimeTypeArray: 0,
      NamedNodeMap: 0,
      NodeList: 1,
      PaintRequestList: 0,
      Plugin: 0,
      PluginArray: 0,
      SVGLengthList: 0,
      SVGNumberList: 0,
      SVGPathSegList: 0,
      SVGPointList: 0,
      SVGStringList: 0,
      SVGTransformList: 0,
      SourceBufferList: 0,
      StyleSheetList: 0,
      TextTrackCueList: 0,
      TextTrackList: 0,
      TouchList: 0
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/dom-token-list-prototype.js
var require_dom_token_list_prototype = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/dom-token-list-prototype.js": function(exports2, module2) {
    "use strict";
    var documentCreateElement = require_document_create_element();
    var classList = documentCreateElement("span").classList;
    var DOMTokenListPrototype = classList && classList.constructor && classList.constructor.prototype;
    module2.exports = DOMTokenListPrototype === Object.prototype ? void 0 : DOMTokenListPrototype;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/array-iteration.js
var require_array_iteration = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/array-iteration.js": function(exports2, module2) {
    "use strict";
    var bind = require_function_bind_context();
    var uncurryThis = require_function_uncurry_this();
    var IndexedObject = require_indexed_object();
    var toObject = require_to_object();
    var lengthOfArrayLike = require_length_of_array_like();
    var arraySpeciesCreate = require_array_species_create();
    var push = uncurryThis([].push);
    var createMethod = function(TYPE) {
      var IS_MAP = TYPE === 1;
      var IS_FILTER = TYPE === 2;
      var IS_SOME = TYPE === 3;
      var IS_EVERY = TYPE === 4;
      var IS_FIND_INDEX = TYPE === 6;
      var IS_FILTER_REJECT = TYPE === 7;
      var NO_HOLES = TYPE === 5 || IS_FIND_INDEX;
      return function($this, callbackfn, that, specificCreate) {
        var O = toObject($this);
        var self2 = IndexedObject(O);
        var length = lengthOfArrayLike(self2);
        var boundFunction = bind(callbackfn, that);
        var index = 0;
        var create = specificCreate || arraySpeciesCreate;
        var target = IS_MAP ? create($this, length) : IS_FILTER || IS_FILTER_REJECT ? create($this, 0) : void 0;
        var value, result;
        for (; length > index; index++) if (NO_HOLES || index in self2) {
          value = self2[index];
          result = boundFunction(value, index, O);
          if (TYPE) {
            if (IS_MAP) target[index] = result;
            else if (result) switch (TYPE) {
              case 3:
                return true;
              // some
              case 5:
                return value;
              // find
              case 6:
                return index;
              // findIndex
              case 2:
                push(target, value);
            }
            else switch (TYPE) {
              case 4:
                return false;
              // every
              case 7:
                push(target, value);
            }
          }
        }
        return IS_FIND_INDEX ? -1 : IS_SOME || IS_EVERY ? IS_EVERY : target;
      };
    };
    module2.exports = {
      // `Array.prototype.forEach` method
      // https://tc39.es/ecma262/#sec-array.prototype.foreach
      forEach: createMethod(0),
      // `Array.prototype.map` method
      // https://tc39.es/ecma262/#sec-array.prototype.map
      map: createMethod(1),
      // `Array.prototype.filter` method
      // https://tc39.es/ecma262/#sec-array.prototype.filter
      filter: createMethod(2),
      // `Array.prototype.some` method
      // https://tc39.es/ecma262/#sec-array.prototype.some
      some: createMethod(3),
      // `Array.prototype.every` method
      // https://tc39.es/ecma262/#sec-array.prototype.every
      every: createMethod(4),
      // `Array.prototype.find` method
      // https://tc39.es/ecma262/#sec-array.prototype.find
      find: createMethod(5),
      // `Array.prototype.findIndex` method
      // https://tc39.es/ecma262/#sec-array.prototype.findIndex
      findIndex: createMethod(6),
      // `Array.prototype.filterReject` method
      // https://github.com/tc39/proposal-array-filtering
      filterReject: createMethod(7)
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/array-method-is-strict.js
var require_array_method_is_strict = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/array-method-is-strict.js": function(exports2, module2) {
    "use strict";
    var fails = require_fails();
    module2.exports = function(METHOD_NAME, argument) {
      var method = [][METHOD_NAME];
      return !!method && fails(function() {
        method.call(null, argument || function() {
          return 1;
        }, 1);
      });
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/array-for-each.js
var require_array_for_each = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/array-for-each.js": function(exports2, module2) {
    "use strict";
    var $forEach = require_array_iteration().forEach;
    var arrayMethodIsStrict = require_array_method_is_strict();
    var STRICT_METHOD = arrayMethodIsStrict("forEach");
    module2.exports = !STRICT_METHOD ? function forEach2(callbackfn) {
      return $forEach(this, callbackfn, arguments.length > 1 ? arguments[1] : void 0);
    } : [].forEach;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/web.dom-collections.for-each.js
var require_web_dom_collections_for_each = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/web.dom-collections.for-each.js": function() {
    "use strict";
    var globalThis2 = require_global_this();
    var DOMIterables = require_dom_iterables();
    var DOMTokenListPrototype = require_dom_token_list_prototype();
    var forEach2 = require_array_for_each();
    var createNonEnumerableProperty = require_create_non_enumerable_property();
    var handlePrototype = function(CollectionPrototype) {
      if (CollectionPrototype && CollectionPrototype.forEach !== forEach2) try {
        createNonEnumerableProperty(CollectionPrototype, "forEach", forEach2);
      } catch (error) {
        CollectionPrototype.forEach = forEach2;
      }
    };
    for (COLLECTION_NAME in DOMIterables) {
      if (DOMIterables[COLLECTION_NAME]) {
        handlePrototype(globalThis2[COLLECTION_NAME] && globalThis2[COLLECTION_NAME].prototype);
      }
    }
    var COLLECTION_NAME;
    handlePrototype(DOMTokenListPrototype);
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/web.dom-collections.iterator.js
var require_web_dom_collections_iterator = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/web.dom-collections.iterator.js": function() {
    "use strict";
    var globalThis2 = require_global_this();
    var DOMIterables = require_dom_iterables();
    var DOMTokenListPrototype = require_dom_token_list_prototype();
    var ArrayIteratorMethods = require_es_array_iterator();
    var createNonEnumerableProperty = require_create_non_enumerable_property();
    var setToStringTag = require_set_to_string_tag();
    var wellKnownSymbol = require_well_known_symbol();
    var ITERATOR = wellKnownSymbol("iterator");
    var ArrayValues = ArrayIteratorMethods.values;
    var handlePrototype = function(CollectionPrototype, COLLECTION_NAME2) {
      if (CollectionPrototype) {
        if (CollectionPrototype[ITERATOR] !== ArrayValues) try {
          createNonEnumerableProperty(CollectionPrototype, ITERATOR, ArrayValues);
        } catch (error) {
          CollectionPrototype[ITERATOR] = ArrayValues;
        }
        setToStringTag(CollectionPrototype, COLLECTION_NAME2, true);
        if (DOMIterables[COLLECTION_NAME2]) for (var METHOD_NAME in ArrayIteratorMethods) {
          if (CollectionPrototype[METHOD_NAME] !== ArrayIteratorMethods[METHOD_NAME]) try {
            createNonEnumerableProperty(CollectionPrototype, METHOD_NAME, ArrayIteratorMethods[METHOD_NAME]);
          } catch (error) {
            CollectionPrototype[METHOD_NAME] = ArrayIteratorMethods[METHOD_NAME];
          }
        }
      }
    };
    for (COLLECTION_NAME in DOMIterables) {
      handlePrototype(globalThis2[COLLECTION_NAME] && globalThis2[COLLECTION_NAME].prototype, COLLECTION_NAME);
    }
    var COLLECTION_NAME;
    handlePrototype(DOMTokenListPrototype, "DOMTokenList");
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.string.iterator.js
var require_es_string_iterator = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.string.iterator.js": function() {
    "use strict";
    var charAt = require_string_multibyte().charAt;
    var toString = require_to_string();
    var InternalStateModule = require_internal_state();
    var defineIterator = require_iterator_define();
    var createIterResultObject = require_create_iter_result_object();
    var STRING_ITERATOR = "String Iterator";
    var setInternalState = InternalStateModule.set;
    var getInternalState = InternalStateModule.getterFor(STRING_ITERATOR);
    defineIterator(String, "String", function(iterated) {
      setInternalState(this, {
        type: STRING_ITERATOR,
        string: toString(iterated),
        index: 0
      });
    }, function next() {
      var state = getInternalState(this);
      var string = state.string;
      var index = state.index;
      var point;
      if (index >= string.length) return createIterResultObject(void 0, true);
      point = charAt(string, index);
      state.index += point.length;
      return createIterResultObject(point, false);
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/url-constructor-detection.js
var require_url_constructor_detection = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/url-constructor-detection.js": function(exports2, module2) {
    "use strict";
    var fails = require_fails();
    var wellKnownSymbol = require_well_known_symbol();
    var DESCRIPTORS = require_descriptors();
    var IS_PURE = require_is_pure();
    var ITERATOR = wellKnownSymbol("iterator");
    module2.exports = !fails(function() {
      var url = new URL("b?a=1&b=2&c=3", "https://a");
      var params = url.searchParams;
      var params2 = new URLSearchParams("a=1&a=2&b=3");
      var result = "";
      url.pathname = "c%20d";
      params.forEach(function(value, key) {
        params["delete"]("b");
        result += key + value;
      });
      params2["delete"]("a", 2);
      params2["delete"]("b", void 0);
      return IS_PURE && (!url.toJSON || !params2.has("a", 1) || params2.has("a", 2) || !params2.has("a", void 0) || params2.has("b")) || !params.size && (IS_PURE || !DESCRIPTORS) || !params.sort || url.href !== "https://a/c%20d?a=1&c=3" || params.get("c") !== "3" || String(new URLSearchParams("?a=1")) !== "a=1" || !params[ITERATOR] || new URL("https://a@b").username !== "a" || new URLSearchParams(new URLSearchParams("a=b")).get("a") !== "b" || new URL("https://\u0442\u0435\u0441\u0442").host !== "xn--e1aybc" || new URL("https://a#\u0431").hash !== "#%D0%B1" || result !== "a1c3" || new URL("https://x", void 0).host !== "x";
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-assign.js
var require_object_assign = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/object-assign.js": function(exports2, module2) {
    "use strict";
    var DESCRIPTORS = require_descriptors();
    var uncurryThis = require_function_uncurry_this();
    var call = require_function_call();
    var fails = require_fails();
    var objectKeys = require_object_keys();
    var getOwnPropertySymbolsModule = require_object_get_own_property_symbols();
    var propertyIsEnumerableModule = require_object_property_is_enumerable();
    var toObject = require_to_object();
    var IndexedObject = require_indexed_object();
    var $assign = Object.assign;
    var defineProperty = Object.defineProperty;
    var concat = uncurryThis([].concat);
    module2.exports = !$assign || fails(function() {
      if (DESCRIPTORS && $assign({ b: 1 }, $assign(defineProperty({}, "a", {
        enumerable: true,
        get: function() {
          defineProperty(this, "b", {
            value: 3,
            enumerable: false
          });
        }
      }), { b: 2 })).b !== 1) return true;
      var A = {};
      var B = {};
      var symbol = Symbol("assign detection");
      var alphabet = "abcdefghijklmnopqrst";
      A[symbol] = 7;
      alphabet.split("").forEach(function(chr) {
        B[chr] = chr;
      });
      return $assign({}, A)[symbol] !== 7 || objectKeys($assign({}, B)).join("") !== alphabet;
    }) ? function assign(target, source) {
      var T = toObject(target);
      var argumentsLength = arguments.length;
      var index = 1;
      var getOwnPropertySymbols = getOwnPropertySymbolsModule.f;
      var propertyIsEnumerable = propertyIsEnumerableModule.f;
      while (argumentsLength > index) {
        var S = IndexedObject(arguments[index++]);
        var keys = getOwnPropertySymbols ? concat(objectKeys(S), getOwnPropertySymbols(S)) : objectKeys(S);
        var length = keys.length;
        var j = 0;
        var key;
        while (length > j) {
          key = keys[j++];
          if (!DESCRIPTORS || call(propertyIsEnumerable, S, key)) T[key] = S[key];
        }
      }
      return T;
    } : $assign;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/call-with-safe-iteration-closing.js
var require_call_with_safe_iteration_closing = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/call-with-safe-iteration-closing.js": function(exports2, module2) {
    "use strict";
    var anObject = require_an_object();
    var iteratorClose = require_iterator_close();
    module2.exports = function(iterator, fn, value, ENTRIES) {
      try {
        return ENTRIES ? fn(anObject(value)[0], value[1]) : fn(value);
      } catch (error) {
        iteratorClose(iterator, "throw", error);
      }
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/array-from.js
var require_array_from = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/array-from.js": function(exports2, module2) {
    "use strict";
    var bind = require_function_bind_context();
    var call = require_function_call();
    var toObject = require_to_object();
    var callWithSafeIterationClosing = require_call_with_safe_iteration_closing();
    var isArrayIteratorMethod = require_is_array_iterator_method();
    var isConstructor = require_is_constructor();
    var lengthOfArrayLike = require_length_of_array_like();
    var createProperty = require_create_property();
    var getIterator = require_get_iterator();
    var getIteratorMethod = require_get_iterator_method();
    var $Array = Array;
    module2.exports = function from(arrayLike) {
      var O = toObject(arrayLike);
      var IS_CONSTRUCTOR = isConstructor(this);
      var argumentsLength = arguments.length;
      var mapfn = argumentsLength > 1 ? arguments[1] : void 0;
      var mapping = mapfn !== void 0;
      if (mapping) mapfn = bind(mapfn, argumentsLength > 2 ? arguments[2] : void 0);
      var iteratorMethod = getIteratorMethod(O);
      var index = 0;
      var length, result, step, iterator, next, value;
      if (iteratorMethod && !(this === $Array && isArrayIteratorMethod(iteratorMethod))) {
        result = IS_CONSTRUCTOR ? new this() : [];
        iterator = getIterator(O, iteratorMethod);
        next = iterator.next;
        for (; !(step = call(next, iterator)).done; index++) {
          value = mapping ? callWithSafeIterationClosing(iterator, mapfn, [step.value, index], true) : step.value;
          createProperty(result, index, value);
        }
      } else {
        length = lengthOfArrayLike(O);
        result = IS_CONSTRUCTOR ? new this(length) : $Array(length);
        for (; length > index; index++) {
          value = mapping ? mapfn(O[index], index) : O[index];
          createProperty(result, index, value);
        }
      }
      result.length = index;
      return result;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/string-punycode-to-ascii.js
var require_string_punycode_to_ascii = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/string-punycode-to-ascii.js": function(exports2, module2) {
    "use strict";
    var uncurryThis = require_function_uncurry_this();
    var maxInt = 2147483647;
    var base = 36;
    var tMin = 1;
    var tMax = 26;
    var skew = 38;
    var damp = 700;
    var initialBias = 72;
    var initialN = 128;
    var delimiter = "-";
    var regexNonASCII = /[^\0-\u007E]/;
    var regexSeparators = /[.\u3002\uFF0E\uFF61]/g;
    var OVERFLOW_ERROR = "Overflow: input needs wider integers to process";
    var baseMinusTMin = base - tMin;
    var $RangeError = RangeError;
    var exec = uncurryThis(regexSeparators.exec);
    var floor = Math.floor;
    var fromCharCode = String.fromCharCode;
    var charCodeAt = uncurryThis("".charCodeAt);
    var join = uncurryThis([].join);
    var push = uncurryThis([].push);
    var replace = uncurryThis("".replace);
    var split = uncurryThis("".split);
    var toLowerCase = uncurryThis("".toLowerCase);
    var ucs2decode = function(string) {
      var output = [];
      var counter = 0;
      var length = string.length;
      while (counter < length) {
        var value = charCodeAt(string, counter++);
        if (value >= 55296 && value <= 56319 && counter < length) {
          var extra = charCodeAt(string, counter++);
          if ((extra & 64512) === 56320) {
            push(output, ((value & 1023) << 10) + (extra & 1023) + 65536);
          } else {
            push(output, value);
            counter--;
          }
        } else {
          push(output, value);
        }
      }
      return output;
    };
    var digitToBasic = function(digit) {
      return digit + 22 + 75 * (digit < 26);
    };
    var adapt = function(delta, numPoints, firstTime) {
      var k = 0;
      delta = firstTime ? floor(delta / damp) : delta >> 1;
      delta += floor(delta / numPoints);
      while (delta > baseMinusTMin * tMax >> 1) {
        delta = floor(delta / baseMinusTMin);
        k += base;
      }
      return floor(k + (baseMinusTMin + 1) * delta / (delta + skew));
    };
    var encode = function(input) {
      var output = [];
      input = ucs2decode(input);
      var inputLength = input.length;
      var n = initialN;
      var delta = 0;
      var bias = initialBias;
      var i, currentValue;
      for (i = 0; i < input.length; i++) {
        currentValue = input[i];
        if (currentValue < 128) {
          push(output, fromCharCode(currentValue));
        }
      }
      var basicLength = output.length;
      var handledCPCount = basicLength;
      if (basicLength) {
        push(output, delimiter);
      }
      while (handledCPCount < inputLength) {
        var m = maxInt;
        for (i = 0; i < input.length; i++) {
          currentValue = input[i];
          if (currentValue >= n && currentValue < m) {
            m = currentValue;
          }
        }
        var handledCPCountPlusOne = handledCPCount + 1;
        if (m - n > floor((maxInt - delta) / handledCPCountPlusOne)) {
          throw new $RangeError(OVERFLOW_ERROR);
        }
        delta += (m - n) * handledCPCountPlusOne;
        n = m;
        for (i = 0; i < input.length; i++) {
          currentValue = input[i];
          if (currentValue < n && ++delta > maxInt) {
            throw new $RangeError(OVERFLOW_ERROR);
          }
          if (currentValue === n) {
            var q = delta;
            var k = base;
            while (true) {
              var t = k <= bias ? tMin : k >= bias + tMax ? tMax : k - bias;
              if (q < t) break;
              var qMinusT = q - t;
              var baseMinusT = base - t;
              push(output, fromCharCode(digitToBasic(t + qMinusT % baseMinusT)));
              q = floor(qMinusT / baseMinusT);
              k += base;
            }
            push(output, fromCharCode(digitToBasic(q)));
            bias = adapt(delta, handledCPCountPlusOne, handledCPCount === basicLength);
            delta = 0;
            handledCPCount++;
          }
        }
        delta++;
        n++;
      }
      return join(output, "");
    };
    module2.exports = function(input) {
      var encoded = [];
      var labels = split(replace(toLowerCase(input), regexSeparators, "."), ".");
      var i, label;
      for (i = 0; i < labels.length; i++) {
        label = labels[i];
        push(encoded, exec(regexNonASCII, label) ? "xn--" + encode(label) : label);
      }
      return join(encoded, ".");
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.string.from-code-point.js
var require_es_string_from_code_point = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/es.string.from-code-point.js": function() {
    "use strict";
    var $ = require_export();
    var uncurryThis = require_function_uncurry_this();
    var toAbsoluteIndex = require_to_absolute_index();
    var $RangeError = RangeError;
    var fromCharCode = String.fromCharCode;
    var $fromCodePoint = String.fromCodePoint;
    var join = uncurryThis([].join);
    var INCORRECT_LENGTH = !!$fromCodePoint && $fromCodePoint.length !== 1;
    $({ target: "String", stat: true, arity: 1, forced: INCORRECT_LENGTH }, {
      // eslint-disable-next-line no-unused-vars -- required for `.length`
      fromCodePoint: function fromCodePoint(x) {
        var elements = [];
        var length = arguments.length;
        var i = 0;
        var code;
        while (length > i) {
          code = +arguments[i++];
          if (toAbsoluteIndex(code, 1114111) !== code) throw new $RangeError(code + " is not a valid code point");
          elements[i] = code < 65536 ? fromCharCode(code) : fromCharCode(((code -= 65536) >> 10) + 55296, code % 1024 + 56320);
        }
        return join(elements, "");
      }
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/define-built-ins.js
var require_define_built_ins = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/define-built-ins.js": function(exports2, module2) {
    "use strict";
    var defineBuiltIn = require_define_built_in();
    module2.exports = function(target, src, options) {
      for (var key in src) defineBuiltIn(target, key, src[key], options);
      return target;
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/array-sort.js
var require_array_sort = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/internals/array-sort.js": function(exports2, module2) {
    "use strict";
    var arraySlice = require_array_slice();
    var floor = Math.floor;
    var sort = function(array, comparefn) {
      var length = array.length;
      if (length < 8) {
        var i = 1;
        var element, j;
        while (i < length) {
          j = i;
          element = array[i];
          while (j && comparefn(array[j - 1], element) > 0) {
            array[j] = array[--j];
          }
          if (j !== i++) array[j] = element;
        }
      } else {
        var middle = floor(length / 2);
        var left = sort(arraySlice(array, 0, middle), comparefn);
        var right = sort(arraySlice(array, middle), comparefn);
        var llength = left.length;
        var rlength = right.length;
        var lindex = 0;
        var rindex = 0;
        while (lindex < llength || rindex < rlength) {
          array[lindex + rindex] = lindex < llength && rindex < rlength ? comparefn(left[lindex], right[rindex]) <= 0 ? left[lindex++] : right[rindex++] : lindex < llength ? left[lindex++] : right[rindex++];
        }
      }
      return array;
    };
    module2.exports = sort;
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/web.url-search-params.constructor.js
var require_web_url_search_params_constructor = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/web.url-search-params.constructor.js": function(exports2, module2) {
    "use strict";
    require_es_array_iterator();
    require_es_string_from_code_point();
    var $ = require_export();
    var globalThis2 = require_global_this();
    var safeGetBuiltIn = require_safe_get_built_in();
    var getBuiltIn = require_get_built_in();
    var call = require_function_call();
    var uncurryThis = require_function_uncurry_this();
    var DESCRIPTORS = require_descriptors();
    var USE_NATIVE_URL = require_url_constructor_detection();
    var defineBuiltIn = require_define_built_in();
    var defineBuiltInAccessor = require_define_built_in_accessor();
    var defineBuiltIns = require_define_built_ins();
    var setToStringTag = require_set_to_string_tag();
    var createIteratorConstructor = require_iterator_create_constructor();
    var InternalStateModule = require_internal_state();
    var anInstance = require_an_instance();
    var isCallable = require_is_callable();
    var hasOwn = require_has_own_property();
    var bind = require_function_bind_context();
    var classof = require_classof();
    var anObject = require_an_object();
    var isObject = require_is_object();
    var $toString = require_to_string();
    var create = require_object_create();
    var createPropertyDescriptor = require_create_property_descriptor();
    var getIterator = require_get_iterator();
    var getIteratorMethod = require_get_iterator_method();
    var createIterResultObject = require_create_iter_result_object();
    var validateArgumentsLength = require_validate_arguments_length();
    var wellKnownSymbol = require_well_known_symbol();
    var arraySort = require_array_sort();
    var ITERATOR = wellKnownSymbol("iterator");
    var URL_SEARCH_PARAMS = "URLSearchParams";
    var URL_SEARCH_PARAMS_ITERATOR = URL_SEARCH_PARAMS + "Iterator";
    var setInternalState = InternalStateModule.set;
    var getInternalParamsState = InternalStateModule.getterFor(URL_SEARCH_PARAMS);
    var getInternalIteratorState = InternalStateModule.getterFor(URL_SEARCH_PARAMS_ITERATOR);
    var nativeFetch = safeGetBuiltIn("fetch");
    var NativeRequest = safeGetBuiltIn("Request");
    var Headers = safeGetBuiltIn("Headers");
    var RequestPrototype = NativeRequest && NativeRequest.prototype;
    var HeadersPrototype = Headers && Headers.prototype;
    var TypeError2 = globalThis2.TypeError;
    var encodeURIComponent2 = globalThis2.encodeURIComponent;
    var fromCharCode = String.fromCharCode;
    var fromCodePoint = getBuiltIn("String", "fromCodePoint");
    var $parseInt = parseInt;
    var charAt = uncurryThis("".charAt);
    var join = uncurryThis([].join);
    var push = uncurryThis([].push);
    var replace = uncurryThis("".replace);
    var shift = uncurryThis([].shift);
    var splice = uncurryThis([].splice);
    var split = uncurryThis("".split);
    var stringSlice = uncurryThis("".slice);
    var exec = uncurryThis(/./.exec);
    var plus = /\+/g;
    var FALLBACK_REPLACER = "\uFFFD";
    var VALID_HEX = /^[0-9a-f]+$/i;
    var parseHexOctet = function(string, start) {
      var substr = stringSlice(string, start, start + 2);
      if (!exec(VALID_HEX, substr)) return NaN;
      return $parseInt(substr, 16);
    };
    var getLeadingOnes = function(octet) {
      var count = 0;
      for (var mask = 128; mask > 0 && (octet & mask) !== 0; mask >>= 1) {
        count++;
      }
      return count;
    };
    var utf8Decode = function(octets) {
      var codePoint = null;
      switch (octets.length) {
        case 1:
          codePoint = octets[0];
          break;
        case 2:
          codePoint = (octets[0] & 31) << 6 | octets[1] & 63;
          break;
        case 3:
          codePoint = (octets[0] & 15) << 12 | (octets[1] & 63) << 6 | octets[2] & 63;
          break;
        case 4:
          codePoint = (octets[0] & 7) << 18 | (octets[1] & 63) << 12 | (octets[2] & 63) << 6 | octets[3] & 63;
          break;
      }
      return codePoint > 1114111 ? null : codePoint;
    };
    var decode = function(input) {
      input = replace(input, plus, " ");
      var length = input.length;
      var result = "";
      var i = 0;
      while (i < length) {
        var decodedChar = charAt(input, i);
        if (decodedChar === "%") {
          if (charAt(input, i + 1) === "%" || i + 3 > length) {
            result += "%";
            i++;
            continue;
          }
          var octet = parseHexOctet(input, i + 1);
          if (octet !== octet) {
            result += decodedChar;
            i++;
            continue;
          }
          i += 2;
          var byteSequenceLength = getLeadingOnes(octet);
          if (byteSequenceLength === 0) {
            decodedChar = fromCharCode(octet);
          } else {
            if (byteSequenceLength === 1 || byteSequenceLength > 4) {
              result += FALLBACK_REPLACER;
              i++;
              continue;
            }
            var octets = [octet];
            var sequenceIndex = 1;
            while (sequenceIndex < byteSequenceLength) {
              i++;
              if (i + 3 > length || charAt(input, i) !== "%") break;
              var nextByte = parseHexOctet(input, i + 1);
              if (nextByte !== nextByte) {
                i += 3;
                break;
              }
              if (nextByte > 191 || nextByte < 128) break;
              push(octets, nextByte);
              i += 2;
              sequenceIndex++;
            }
            if (octets.length !== byteSequenceLength) {
              result += FALLBACK_REPLACER;
              continue;
            }
            var codePoint = utf8Decode(octets);
            if (codePoint === null) {
              result += FALLBACK_REPLACER;
            } else {
              decodedChar = fromCodePoint(codePoint);
            }
          }
        }
        result += decodedChar;
        i++;
      }
      return result;
    };
    var find2 = /[!'()~]|%20/g;
    var replacements = {
      "!": "%21",
      "'": "%27",
      "(": "%28",
      ")": "%29",
      "~": "%7E",
      "%20": "+"
    };
    var replacer = function(match) {
      return replacements[match];
    };
    var serialize = function(it) {
      return replace(encodeURIComponent2(it), find2, replacer);
    };
    var URLSearchParamsIterator = createIteratorConstructor(function Iterator(params, kind) {
      setInternalState(this, {
        type: URL_SEARCH_PARAMS_ITERATOR,
        target: getInternalParamsState(params).entries,
        index: 0,
        kind: kind
      });
    }, URL_SEARCH_PARAMS, function next() {
      var state = getInternalIteratorState(this);
      var target = state.target;
      var index = state.index++;
      if (!target || index >= target.length) {
        state.target = null;
        return createIterResultObject(void 0, true);
      }
      var entry = target[index];
      switch (state.kind) {
        case "keys":
          return createIterResultObject(entry.key, false);
        case "values":
          return createIterResultObject(entry.value, false);
      }
      return createIterResultObject([entry.key, entry.value], false);
    }, true);
    var URLSearchParamsState = function(init2) {
      this.entries = [];
      this.url = null;
      if (init2 !== void 0) {
        if (isObject(init2)) this.parseObject(init2);
        else this.parseQuery(typeof init2 == "string" ? charAt(init2, 0) === "?" ? stringSlice(init2, 1) : init2 : $toString(init2));
      }
    };
    URLSearchParamsState.prototype = {
      type: URL_SEARCH_PARAMS,
      bindURL: function(url) {
        this.url = url;
        this.update();
      },
      parseObject: function(object) {
        var entries = this.entries;
        var iteratorMethod = getIteratorMethod(object);
        var iterator, next, step, entryIterator, entryNext, first, second;
        if (iteratorMethod) {
          iterator = getIterator(object, iteratorMethod);
          next = iterator.next;
          while (!(step = call(next, iterator)).done) {
            entryIterator = getIterator(anObject(step.value));
            entryNext = entryIterator.next;
            if ((first = call(entryNext, entryIterator)).done || (second = call(entryNext, entryIterator)).done || !call(entryNext, entryIterator).done) throw new TypeError2("Expected sequence with length 2");
            push(entries, { key: $toString(first.value), value: $toString(second.value) });
          }
        } else for (var key in object) if (hasOwn(object, key)) {
          push(entries, { key: key, value: $toString(object[key]) });
        }
      },
      parseQuery: function(query) {
        if (query) {
          var entries = this.entries;
          var attributes = split(query, "&");
          var index = 0;
          var attribute, entry;
          while (index < attributes.length) {
            attribute = attributes[index++];
            if (attribute.length) {
              entry = split(attribute, "=");
              push(entries, {
                key: decode(shift(entry)),
                value: decode(join(entry, "="))
              });
            }
          }
        }
      },
      serialize: function() {
        var entries = this.entries;
        var result = [];
        var index = 0;
        var entry;
        while (index < entries.length) {
          entry = entries[index++];
          push(result, serialize(entry.key) + "=" + serialize(entry.value));
        }
        return join(result, "&");
      },
      update: function() {
        this.entries.length = 0;
        this.parseQuery(this.url.query);
      },
      updateURL: function() {
        if (this.url) this.url.update();
      }
    };
    var URLSearchParamsConstructor = function URLSearchParams2() {
      anInstance(this, URLSearchParamsPrototype);
      var init2 = arguments.length > 0 ? arguments[0] : void 0;
      var state = setInternalState(this, new URLSearchParamsState(init2));
      if (!DESCRIPTORS) this.size = state.entries.length;
    };
    var URLSearchParamsPrototype = URLSearchParamsConstructor.prototype;
    defineBuiltIns(URLSearchParamsPrototype, {
      // `URLSearchParams.prototype.append` method
      // https://url.spec.whatwg.org/#dom-urlsearchparams-append
      append: function append(name, value) {
        var state = getInternalParamsState(this);
        validateArgumentsLength(arguments.length, 2);
        push(state.entries, { key: $toString(name), value: $toString(value) });
        if (!DESCRIPTORS) this.length++;
        state.updateURL();
      },
      // `URLSearchParams.prototype.delete` method
      // https://url.spec.whatwg.org/#dom-urlsearchparams-delete
      "delete": function(name) {
        var state = getInternalParamsState(this);
        var length = validateArgumentsLength(arguments.length, 1);
        var entries = state.entries;
        var key = $toString(name);
        var $value = length < 2 ? void 0 : arguments[1];
        var value = $value === void 0 ? $value : $toString($value);
        var index = 0;
        while (index < entries.length) {
          var entry = entries[index];
          if (entry.key === key && (value === void 0 || entry.value === value)) {
            splice(entries, index, 1);
            if (value !== void 0) break;
          } else index++;
        }
        if (!DESCRIPTORS) this.size = entries.length;
        state.updateURL();
      },
      // `URLSearchParams.prototype.get` method
      // https://url.spec.whatwg.org/#dom-urlsearchparams-get
      get: function get(name) {
        var entries = getInternalParamsState(this).entries;
        validateArgumentsLength(arguments.length, 1);
        var key = $toString(name);
        var index = 0;
        for (; index < entries.length; index++) {
          if (entries[index].key === key) return entries[index].value;
        }
        return null;
      },
      // `URLSearchParams.prototype.getAll` method
      // https://url.spec.whatwg.org/#dom-urlsearchparams-getall
      getAll: function getAll(name) {
        var entries = getInternalParamsState(this).entries;
        validateArgumentsLength(arguments.length, 1);
        var key = $toString(name);
        var result = [];
        var index = 0;
        for (; index < entries.length; index++) {
          if (entries[index].key === key) push(result, entries[index].value);
        }
        return result;
      },
      // `URLSearchParams.prototype.has` method
      // https://url.spec.whatwg.org/#dom-urlsearchparams-has
      has: function has(name) {
        var entries = getInternalParamsState(this).entries;
        var length = validateArgumentsLength(arguments.length, 1);
        var key = $toString(name);
        var $value = length < 2 ? void 0 : arguments[1];
        var value = $value === void 0 ? $value : $toString($value);
        var index = 0;
        while (index < entries.length) {
          var entry = entries[index++];
          if (entry.key === key && (value === void 0 || entry.value === value)) return true;
        }
        return false;
      },
      // `URLSearchParams.prototype.set` method
      // https://url.spec.whatwg.org/#dom-urlsearchparams-set
      set: function set(name, value) {
        var state = getInternalParamsState(this);
        validateArgumentsLength(arguments.length, 1);
        var entries = state.entries;
        var found = false;
        var key = $toString(name);
        var val = $toString(value);
        var index = 0;
        var entry;
        for (; index < entries.length; index++) {
          entry = entries[index];
          if (entry.key === key) {
            if (found) splice(entries, index--, 1);
            else {
              found = true;
              entry.value = val;
            }
          }
        }
        if (!found) push(entries, { key: key, value: val });
        if (!DESCRIPTORS) this.size = entries.length;
        state.updateURL();
      },
      // `URLSearchParams.prototype.sort` method
      // https://url.spec.whatwg.org/#dom-urlsearchparams-sort
      sort: function sort() {
        var state = getInternalParamsState(this);
        arraySort(state.entries, function(a, b) {
          return a.key > b.key ? 1 : -1;
        });
        state.updateURL();
      },
      // `URLSearchParams.prototype.forEach` method
      forEach: function forEach2(callback) {
        var entries = getInternalParamsState(this).entries;
        var boundFunction = bind(callback, arguments.length > 1 ? arguments[1] : void 0);
        var index = 0;
        var entry;
        while (index < entries.length) {
          entry = entries[index++];
          boundFunction(entry.value, entry.key, this);
        }
      },
      // `URLSearchParams.prototype.keys` method
      keys: function keys() {
        return new URLSearchParamsIterator(this, "keys");
      },
      // `URLSearchParams.prototype.values` method
      values: function values() {
        return new URLSearchParamsIterator(this, "values");
      },
      // `URLSearchParams.prototype.entries` method
      entries: function entries() {
        return new URLSearchParamsIterator(this, "entries");
      }
    }, { enumerable: true });
    defineBuiltIn(URLSearchParamsPrototype, ITERATOR, URLSearchParamsPrototype.entries, { name: "entries" });
    defineBuiltIn(URLSearchParamsPrototype, "toString", function toString() {
      return getInternalParamsState(this).serialize();
    }, { enumerable: true });
    if (DESCRIPTORS) defineBuiltInAccessor(URLSearchParamsPrototype, "size", {
      get: function size() {
        return getInternalParamsState(this).entries.length;
      },
      configurable: true,
      enumerable: true
    });
    setToStringTag(URLSearchParamsConstructor, URL_SEARCH_PARAMS);
    $({ global: true, constructor: true, forced: !USE_NATIVE_URL }, {
      URLSearchParams: URLSearchParamsConstructor
    });
    if (!USE_NATIVE_URL && isCallable(Headers)) {
      headersHas = uncurryThis(HeadersPrototype.has);
      headersSet = uncurryThis(HeadersPrototype.set);
      wrapRequestOptions = function(init2) {
        if (isObject(init2)) {
          var body = init2.body;
          var headers;
          if (classof(body) === URL_SEARCH_PARAMS) {
            headers = init2.headers ? new Headers(init2.headers) : new Headers();
            if (!headersHas(headers, "content-type")) {
              headersSet(headers, "content-type", "application/x-www-form-urlencoded;charset=UTF-8");
            }
            return create(init2, {
              body: createPropertyDescriptor(0, $toString(body)),
              headers: createPropertyDescriptor(0, headers)
            });
          }
        }
        return init2;
      };
      if (isCallable(nativeFetch)) {
        $({ global: true, enumerable: true, dontCallGetSet: true, forced: true }, {
          fetch: function fetch(input) {
            return nativeFetch(input, arguments.length > 1 ? wrapRequestOptions(arguments[1]) : {});
          }
        });
      }
      if (isCallable(NativeRequest)) {
        RequestConstructor = function Request(input) {
          anInstance(this, RequestPrototype);
          return new NativeRequest(input, arguments.length > 1 ? wrapRequestOptions(arguments[1]) : {});
        };
        RequestPrototype.constructor = RequestConstructor;
        RequestConstructor.prototype = RequestPrototype;
        $({ global: true, constructor: true, dontCallGetSet: true, forced: true }, {
          Request: RequestConstructor
        });
      }
    }
    var headersHas;
    var headersSet;
    var wrapRequestOptions;
    var RequestConstructor;
    module2.exports = {
      URLSearchParams: URLSearchParamsConstructor,
      getState: getInternalParamsState
    };
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/web.url.constructor.js
var require_web_url_constructor = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/web.url.constructor.js": function() {
    "use strict";
    require_es_string_iterator();
    var $ = require_export();
    var DESCRIPTORS = require_descriptors();
    var USE_NATIVE_URL = require_url_constructor_detection();
    var globalThis2 = require_global_this();
    var bind = require_function_bind_context();
    var uncurryThis = require_function_uncurry_this();
    var defineBuiltIn = require_define_built_in();
    var defineBuiltInAccessor = require_define_built_in_accessor();
    var anInstance = require_an_instance();
    var hasOwn = require_has_own_property();
    var assign = require_object_assign();
    var arrayFrom = require_array_from();
    var arraySlice = require_array_slice();
    var codeAt = require_string_multibyte().codeAt;
    var toASCII = require_string_punycode_to_ascii();
    var $toString = require_to_string();
    var setToStringTag = require_set_to_string_tag();
    var validateArgumentsLength = require_validate_arguments_length();
    var URLSearchParamsModule = require_web_url_search_params_constructor();
    var InternalStateModule = require_internal_state();
    var setInternalState = InternalStateModule.set;
    var getInternalURLState = InternalStateModule.getterFor("URL");
    var URLSearchParams2 = URLSearchParamsModule.URLSearchParams;
    var getInternalSearchParamsState = URLSearchParamsModule.getState;
    var NativeURL = globalThis2.URL;
    var TypeError2 = globalThis2.TypeError;
    var parseInt2 = globalThis2.parseInt;
    var floor = Math.floor;
    var pow = Math.pow;
    var charAt = uncurryThis("".charAt);
    var exec = uncurryThis(/./.exec);
    var join = uncurryThis([].join);
    var numberToString = uncurryThis(1 .toString);
    var pop = uncurryThis([].pop);
    var push = uncurryThis([].push);
    var replace = uncurryThis("".replace);
    var shift = uncurryThis([].shift);
    var split = uncurryThis("".split);
    var stringSlice = uncurryThis("".slice);
    var toLowerCase = uncurryThis("".toLowerCase);
    var unshift = uncurryThis([].unshift);
    var INVALID_AUTHORITY = "Invalid authority";
    var INVALID_SCHEME = "Invalid scheme";
    var INVALID_HOST = "Invalid host";
    var INVALID_PORT = "Invalid port";
    var ALPHA = /[a-z]/i;
    var ALPHANUMERIC = /[\d+-.a-z]/i;
    var DIGIT = /\d/;
    var HEX_START = /^0x/i;
    var OCT = /^[0-7]+$/;
    var DEC = /^\d+$/;
    var HEX = /^[\da-f]+$/i;
    var FORBIDDEN_HOST_CODE_POINT = /[\0\t\n\r #%/:<>?@[\\\]^|]/;
    var FORBIDDEN_HOST_CODE_POINT_EXCLUDING_PERCENT = /[\0\t\n\r #/:<>?@[\\\]^|]/;
    var LEADING_C0_CONTROL_OR_SPACE = /^[\u0000-\u0020]+/;
    var TRAILING_C0_CONTROL_OR_SPACE = /(^|[^\u0000-\u0020])[\u0000-\u0020]+$/;
    var TAB_AND_NEW_LINE = /[\t\n\r]/g;
    var EOF;
    var parseIPv4 = function(input) {
      var parts = split(input, ".");
      var partsLength, numbers, index, part, radix, number, ipv4;
      if (parts.length && parts[parts.length - 1] === "") {
        parts.length--;
      }
      partsLength = parts.length;
      if (partsLength > 4) return input;
      numbers = [];
      for (index = 0; index < partsLength; index++) {
        part = parts[index];
        if (part === "") return input;
        radix = 10;
        if (part.length > 1 && charAt(part, 0) === "0") {
          radix = exec(HEX_START, part) ? 16 : 8;
          part = stringSlice(part, radix === 8 ? 1 : 2);
        }
        if (part === "") {
          number = 0;
        } else {
          if (!exec(radix === 10 ? DEC : radix === 8 ? OCT : HEX, part)) return input;
          number = parseInt2(part, radix);
        }
        push(numbers, number);
      }
      for (index = 0; index < partsLength; index++) {
        number = numbers[index];
        if (index === partsLength - 1) {
          if (number >= pow(256, 5 - partsLength)) return null;
        } else if (number > 255) return null;
      }
      ipv4 = pop(numbers);
      for (index = 0; index < numbers.length; index++) {
        ipv4 += numbers[index] * pow(256, 3 - index);
      }
      return ipv4;
    };
    var parseIPv6 = function(input) {
      var address = [0, 0, 0, 0, 0, 0, 0, 0];
      var pieceIndex = 0;
      var compress = null;
      var pointer = 0;
      var value, length, numbersSeen, ipv4Piece, number, swaps, swap2;
      var chr = function() {
        return charAt(input, pointer);
      };
      if (chr() === ":") {
        if (charAt(input, 1) !== ":") return;
        pointer += 2;
        pieceIndex++;
        compress = pieceIndex;
      }
      while (chr()) {
        if (pieceIndex === 8) return;
        if (chr() === ":") {
          if (compress !== null) return;
          pointer++;
          pieceIndex++;
          compress = pieceIndex;
          continue;
        }
        value = length = 0;
        while (length < 4 && exec(HEX, chr())) {
          value = value * 16 + parseInt2(chr(), 16);
          pointer++;
          length++;
        }
        if (chr() === ".") {
          if (length === 0) return;
          pointer -= length;
          if (pieceIndex > 6) return;
          numbersSeen = 0;
          while (chr()) {
            ipv4Piece = null;
            if (numbersSeen > 0) {
              if (chr() === "." && numbersSeen < 4) pointer++;
              else return;
            }
            if (!exec(DIGIT, chr())) return;
            while (exec(DIGIT, chr())) {
              number = parseInt2(chr(), 10);
              if (ipv4Piece === null) ipv4Piece = number;
              else if (ipv4Piece === 0) return;
              else ipv4Piece = ipv4Piece * 10 + number;
              if (ipv4Piece > 255) return;
              pointer++;
            }
            address[pieceIndex] = address[pieceIndex] * 256 + ipv4Piece;
            numbersSeen++;
            if (numbersSeen === 2 || numbersSeen === 4) pieceIndex++;
          }
          if (numbersSeen !== 4) return;
          break;
        } else if (chr() === ":") {
          pointer++;
          if (!chr()) return;
        } else if (chr()) return;
        address[pieceIndex++] = value;
      }
      if (compress !== null) {
        swaps = pieceIndex - compress;
        pieceIndex = 7;
        while (pieceIndex !== 0 && swaps > 0) {
          swap2 = address[pieceIndex];
          address[pieceIndex--] = address[compress + swaps - 1];
          address[compress + --swaps] = swap2;
        }
      } else if (pieceIndex !== 8) return;
      return address;
    };
    var findLongestZeroSequence = function(ipv6) {
      var maxIndex = null;
      var maxLength = 1;
      var currStart = null;
      var currLength = 0;
      var index = 0;
      for (; index < 8; index++) {
        if (ipv6[index] !== 0) {
          if (currLength > maxLength) {
            maxIndex = currStart;
            maxLength = currLength;
          }
          currStart = null;
          currLength = 0;
        } else {
          if (currStart === null) currStart = index;
          ++currLength;
        }
      }
      return currLength > maxLength ? currStart : maxIndex;
    };
    var serializeHost = function(host) {
      var result, index, compress, ignore0;
      if (typeof host == "number") {
        result = [];
        for (index = 0; index < 4; index++) {
          unshift(result, host % 256);
          host = floor(host / 256);
        }
        return join(result, ".");
      }
      if (typeof host == "object") {
        result = "";
        compress = findLongestZeroSequence(host);
        for (index = 0; index < 8; index++) {
          if (ignore0 && host[index] === 0) continue;
          if (ignore0) ignore0 = false;
          if (compress === index) {
            result += index ? ":" : "::";
            ignore0 = true;
          } else {
            result += numberToString(host[index], 16);
            if (index < 7) result += ":";
          }
        }
        return "[" + result + "]";
      }
      return host;
    };
    var C0ControlPercentEncodeSet = {};
    var fragmentPercentEncodeSet = assign({}, C0ControlPercentEncodeSet, {
      " ": 1,
      '"': 1,
      "<": 1,
      ">": 1,
      "`": 1
    });
    var pathPercentEncodeSet = assign({}, fragmentPercentEncodeSet, {
      "#": 1,
      "?": 1,
      "{": 1,
      "}": 1
    });
    var userinfoPercentEncodeSet = assign({}, pathPercentEncodeSet, {
      "/": 1,
      ":": 1,
      ";": 1,
      "=": 1,
      "@": 1,
      "[": 1,
      "\\": 1,
      "]": 1,
      "^": 1,
      "|": 1
    });
    var percentEncode = function(chr, set) {
      var code = codeAt(chr, 0);
      return code > 32 && code < 127 && !hasOwn(set, chr) ? chr : encodeURIComponent(chr);
    };
    var specialSchemes = {
      ftp: 21,
      file: null,
      http: 80,
      https: 443,
      ws: 80,
      wss: 443
    };
    var isWindowsDriveLetter = function(string, normalized) {
      var second;
      return string.length === 2 && exec(ALPHA, charAt(string, 0)) && ((second = charAt(string, 1)) === ":" || !normalized && second === "|");
    };
    var startsWithWindowsDriveLetter = function(string) {
      var third;
      return string.length > 1 && isWindowsDriveLetter(stringSlice(string, 0, 2)) && (string.length === 2 || ((third = charAt(string, 2)) === "/" || third === "\\" || third === "?" || third === "#"));
    };
    var isSingleDot = function(segment) {
      return segment === "." || toLowerCase(segment) === "%2e";
    };
    var isDoubleDot = function(segment) {
      segment = toLowerCase(segment);
      return segment === ".." || segment === "%2e." || segment === ".%2e" || segment === "%2e%2e";
    };
    var SCHEME_START = {};
    var SCHEME = {};
    var NO_SCHEME = {};
    var SPECIAL_RELATIVE_OR_AUTHORITY = {};
    var PATH_OR_AUTHORITY = {};
    var RELATIVE = {};
    var RELATIVE_SLASH = {};
    var SPECIAL_AUTHORITY_SLASHES = {};
    var SPECIAL_AUTHORITY_IGNORE_SLASHES = {};
    var AUTHORITY = {};
    var HOST = {};
    var HOSTNAME = {};
    var PORT = {};
    var FILE = {};
    var FILE_SLASH = {};
    var FILE_HOST = {};
    var PATH_START = {};
    var PATH = {};
    var CANNOT_BE_A_BASE_URL_PATH = {};
    var QUERY = {};
    var FRAGMENT = {};
    var URLState = function(url, isBase, base) {
      var urlString = $toString(url);
      var baseState, failure, searchParams;
      if (isBase) {
        failure = this.parse(urlString);
        if (failure) throw new TypeError2(failure);
        this.searchParams = null;
      } else {
        if (base !== void 0) baseState = new URLState(base, true);
        failure = this.parse(urlString, null, baseState);
        if (failure) throw new TypeError2(failure);
        searchParams = getInternalSearchParamsState(new URLSearchParams2());
        searchParams.bindURL(this);
        this.searchParams = searchParams;
      }
    };
    URLState.prototype = {
      type: "URL",
      // https://url.spec.whatwg.org/#url-parsing
      // eslint-disable-next-line max-statements -- TODO
      parse: function(input, stateOverride, base) {
        var url = this;
        var state = stateOverride || SCHEME_START;
        var pointer = 0;
        var buffer = "";
        var seenAt = false;
        var seenBracket = false;
        var seenPasswordToken = false;
        var codePoints, chr, bufferCodePoints, failure;
        input = $toString(input);
        if (!stateOverride) {
          url.scheme = "";
          url.username = "";
          url.password = "";
          url.host = null;
          url.port = null;
          url.path = [];
          url.query = null;
          url.fragment = null;
          url.cannotBeABaseURL = false;
          input = replace(input, LEADING_C0_CONTROL_OR_SPACE, "");
          input = replace(input, TRAILING_C0_CONTROL_OR_SPACE, "$1");
        }
        input = replace(input, TAB_AND_NEW_LINE, "");
        codePoints = arrayFrom(input);
        while (pointer <= codePoints.length) {
          chr = codePoints[pointer];
          switch (state) {
            case SCHEME_START:
              if (chr && exec(ALPHA, chr)) {
                buffer += toLowerCase(chr);
                state = SCHEME;
              } else if (!stateOverride) {
                state = NO_SCHEME;
                continue;
              } else return INVALID_SCHEME;
              break;
            case SCHEME:
              if (chr && (exec(ALPHANUMERIC, chr) || chr === "+" || chr === "-" || chr === ".")) {
                buffer += toLowerCase(chr);
              } else if (chr === ":") {
                if (stateOverride && (url.isSpecial() !== hasOwn(specialSchemes, buffer) || buffer === "file" && (url.includesCredentials() || url.port !== null) || url.scheme === "file" && !url.host)) return;
                url.scheme = buffer;
                if (stateOverride) {
                  if (url.isSpecial() && specialSchemes[url.scheme] === url.port) url.port = null;
                  return;
                }
                buffer = "";
                if (url.scheme === "file") {
                  state = FILE;
                } else if (url.isSpecial() && base && base.scheme === url.scheme) {
                  state = SPECIAL_RELATIVE_OR_AUTHORITY;
                } else if (url.isSpecial()) {
                  state = SPECIAL_AUTHORITY_SLASHES;
                } else if (codePoints[pointer + 1] === "/") {
                  state = PATH_OR_AUTHORITY;
                  pointer++;
                } else {
                  url.cannotBeABaseURL = true;
                  push(url.path, "");
                  state = CANNOT_BE_A_BASE_URL_PATH;
                }
              } else if (!stateOverride) {
                buffer = "";
                state = NO_SCHEME;
                pointer = 0;
                continue;
              } else return INVALID_SCHEME;
              break;
            case NO_SCHEME:
              if (!base || base.cannotBeABaseURL && chr !== "#") return INVALID_SCHEME;
              if (base.cannotBeABaseURL && chr === "#") {
                url.scheme = base.scheme;
                url.path = arraySlice(base.path);
                url.query = base.query;
                url.fragment = "";
                url.cannotBeABaseURL = true;
                state = FRAGMENT;
                break;
              }
              state = base.scheme === "file" ? FILE : RELATIVE;
              continue;
            case SPECIAL_RELATIVE_OR_AUTHORITY:
              if (chr === "/" && codePoints[pointer + 1] === "/") {
                state = SPECIAL_AUTHORITY_IGNORE_SLASHES;
                pointer++;
              } else {
                state = RELATIVE;
                continue;
              }
              break;
            case PATH_OR_AUTHORITY:
              if (chr === "/") {
                state = AUTHORITY;
                break;
              } else {
                state = PATH;
                continue;
              }
            case RELATIVE:
              url.scheme = base.scheme;
              if (chr === EOF) {
                url.username = base.username;
                url.password = base.password;
                url.host = base.host;
                url.port = base.port;
                url.path = arraySlice(base.path);
                url.query = base.query;
              } else if (chr === "/" || chr === "\\" && url.isSpecial()) {
                state = RELATIVE_SLASH;
              } else if (chr === "?") {
                url.username = base.username;
                url.password = base.password;
                url.host = base.host;
                url.port = base.port;
                url.path = arraySlice(base.path);
                url.query = "";
                state = QUERY;
              } else if (chr === "#") {
                url.username = base.username;
                url.password = base.password;
                url.host = base.host;
                url.port = base.port;
                url.path = arraySlice(base.path);
                url.query = base.query;
                url.fragment = "";
                state = FRAGMENT;
              } else {
                url.username = base.username;
                url.password = base.password;
                url.host = base.host;
                url.port = base.port;
                url.path = arraySlice(base.path);
                url.path.length--;
                state = PATH;
                continue;
              }
              break;
            case RELATIVE_SLASH:
              if (url.isSpecial() && (chr === "/" || chr === "\\")) {
                state = SPECIAL_AUTHORITY_IGNORE_SLASHES;
              } else if (chr === "/") {
                state = AUTHORITY;
              } else {
                url.username = base.username;
                url.password = base.password;
                url.host = base.host;
                url.port = base.port;
                state = PATH;
                continue;
              }
              break;
            case SPECIAL_AUTHORITY_SLASHES:
              state = SPECIAL_AUTHORITY_IGNORE_SLASHES;
              if (chr !== "/" || charAt(buffer, pointer + 1) !== "/") continue;
              pointer++;
              break;
            case SPECIAL_AUTHORITY_IGNORE_SLASHES:
              if (chr !== "/" && chr !== "\\") {
                state = AUTHORITY;
                continue;
              }
              break;
            case AUTHORITY:
              if (chr === "@") {
                if (seenAt) buffer = "%40" + buffer;
                seenAt = true;
                bufferCodePoints = arrayFrom(buffer);
                for (var i = 0; i < bufferCodePoints.length; i++) {
                  var codePoint = bufferCodePoints[i];
                  if (codePoint === ":" && !seenPasswordToken) {
                    seenPasswordToken = true;
                    continue;
                  }
                  var encodedCodePoints = percentEncode(codePoint, userinfoPercentEncodeSet);
                  if (seenPasswordToken) url.password += encodedCodePoints;
                  else url.username += encodedCodePoints;
                }
                buffer = "";
              } else if (chr === EOF || chr === "/" || chr === "?" || chr === "#" || chr === "\\" && url.isSpecial()) {
                if (seenAt && buffer === "") return INVALID_AUTHORITY;
                pointer -= arrayFrom(buffer).length + 1;
                buffer = "";
                state = HOST;
              } else buffer += chr;
              break;
            case HOST:
            case HOSTNAME:
              if (stateOverride && url.scheme === "file") {
                state = FILE_HOST;
                continue;
              } else if (chr === ":" && !seenBracket) {
                if (buffer === "") return INVALID_HOST;
                failure = url.parseHost(buffer);
                if (failure) return failure;
                buffer = "";
                state = PORT;
                if (stateOverride === HOSTNAME) return;
              } else if (chr === EOF || chr === "/" || chr === "?" || chr === "#" || chr === "\\" && url.isSpecial()) {
                if (url.isSpecial() && buffer === "") return INVALID_HOST;
                if (stateOverride && buffer === "" && (url.includesCredentials() || url.port !== null)) return;
                failure = url.parseHost(buffer);
                if (failure) return failure;
                buffer = "";
                state = PATH_START;
                if (stateOverride) return;
                continue;
              } else {
                if (chr === "[") seenBracket = true;
                else if (chr === "]") seenBracket = false;
                buffer += chr;
              }
              break;
            case PORT:
              if (exec(DIGIT, chr)) {
                buffer += chr;
              } else if (chr === EOF || chr === "/" || chr === "?" || chr === "#" || chr === "\\" && url.isSpecial() || stateOverride) {
                if (buffer !== "") {
                  var port = parseInt2(buffer, 10);
                  if (port > 65535) return INVALID_PORT;
                  url.port = url.isSpecial() && port === specialSchemes[url.scheme] ? null : port;
                  buffer = "";
                }
                if (stateOverride) return;
                state = PATH_START;
                continue;
              } else return INVALID_PORT;
              break;
            case FILE:
              url.scheme = "file";
              if (chr === "/" || chr === "\\") state = FILE_SLASH;
              else if (base && base.scheme === "file") {
                switch (chr) {
                  case EOF:
                    url.host = base.host;
                    url.path = arraySlice(base.path);
                    url.query = base.query;
                    break;
                  case "?":
                    url.host = base.host;
                    url.path = arraySlice(base.path);
                    url.query = "";
                    state = QUERY;
                    break;
                  case "#":
                    url.host = base.host;
                    url.path = arraySlice(base.path);
                    url.query = base.query;
                    url.fragment = "";
                    state = FRAGMENT;
                    break;
                  default:
                    if (!startsWithWindowsDriveLetter(join(arraySlice(codePoints, pointer), ""))) {
                      url.host = base.host;
                      url.path = arraySlice(base.path);
                      url.shortenPath();
                    }
                    state = PATH;
                    continue;
                }
              } else {
                state = PATH;
                continue;
              }
              break;
            case FILE_SLASH:
              if (chr === "/" || chr === "\\") {
                state = FILE_HOST;
                break;
              }
              if (base && base.scheme === "file" && !startsWithWindowsDriveLetter(join(arraySlice(codePoints, pointer), ""))) {
                if (isWindowsDriveLetter(base.path[0], true)) push(url.path, base.path[0]);
                else url.host = base.host;
              }
              state = PATH;
              continue;
            case FILE_HOST:
              if (chr === EOF || chr === "/" || chr === "\\" || chr === "?" || chr === "#") {
                if (!stateOverride && isWindowsDriveLetter(buffer)) {
                  state = PATH;
                } else if (buffer === "") {
                  url.host = "";
                  if (stateOverride) return;
                  state = PATH_START;
                } else {
                  failure = url.parseHost(buffer);
                  if (failure) return failure;
                  if (url.host === "localhost") url.host = "";
                  if (stateOverride) return;
                  buffer = "";
                  state = PATH_START;
                }
                continue;
              } else buffer += chr;
              break;
            case PATH_START:
              if (url.isSpecial()) {
                state = PATH;
                if (chr !== "/" && chr !== "\\") continue;
              } else if (!stateOverride && chr === "?") {
                url.query = "";
                state = QUERY;
              } else if (!stateOverride && chr === "#") {
                url.fragment = "";
                state = FRAGMENT;
              } else if (chr !== EOF) {
                state = PATH;
                if (chr !== "/") continue;
              }
              break;
            case PATH:
              if (chr === EOF || chr === "/" || chr === "\\" && url.isSpecial() || !stateOverride && (chr === "?" || chr === "#")) {
                if (isDoubleDot(buffer)) {
                  url.shortenPath();
                  if (chr !== "/" && !(chr === "\\" && url.isSpecial())) {
                    push(url.path, "");
                  }
                } else if (isSingleDot(buffer)) {
                  if (chr !== "/" && !(chr === "\\" && url.isSpecial())) {
                    push(url.path, "");
                  }
                } else {
                  if (url.scheme === "file" && !url.path.length && isWindowsDriveLetter(buffer)) {
                    if (url.host) url.host = "";
                    buffer = charAt(buffer, 0) + ":";
                  }
                  push(url.path, buffer);
                }
                buffer = "";
                if (url.scheme === "file" && (chr === EOF || chr === "?" || chr === "#")) {
                  while (url.path.length > 1 && url.path[0] === "") {
                    shift(url.path);
                  }
                }
                if (chr === "?") {
                  url.query = "";
                  state = QUERY;
                } else if (chr === "#") {
                  url.fragment = "";
                  state = FRAGMENT;
                }
              } else {
                buffer += percentEncode(chr, pathPercentEncodeSet);
              }
              break;
            case CANNOT_BE_A_BASE_URL_PATH:
              if (chr === "?") {
                url.query = "";
                state = QUERY;
              } else if (chr === "#") {
                url.fragment = "";
                state = FRAGMENT;
              } else if (chr !== EOF) {
                url.path[0] += percentEncode(chr, C0ControlPercentEncodeSet);
              }
              break;
            case QUERY:
              if (!stateOverride && chr === "#") {
                url.fragment = "";
                state = FRAGMENT;
              } else if (chr !== EOF) {
                if (chr === "'" && url.isSpecial()) url.query += "%27";
                else if (chr === "#") url.query += "%23";
                else url.query += percentEncode(chr, C0ControlPercentEncodeSet);
              }
              break;
            case FRAGMENT:
              if (chr !== EOF) url.fragment += percentEncode(chr, fragmentPercentEncodeSet);
              break;
          }
          pointer++;
        }
      },
      // https://url.spec.whatwg.org/#host-parsing
      parseHost: function(input) {
        var result, codePoints, index;
        if (charAt(input, 0) === "[") {
          if (charAt(input, input.length - 1) !== "]") return INVALID_HOST;
          result = parseIPv6(stringSlice(input, 1, -1));
          if (!result) return INVALID_HOST;
          this.host = result;
        } else if (!this.isSpecial()) {
          if (exec(FORBIDDEN_HOST_CODE_POINT_EXCLUDING_PERCENT, input)) return INVALID_HOST;
          result = "";
          codePoints = arrayFrom(input);
          for (index = 0; index < codePoints.length; index++) {
            result += percentEncode(codePoints[index], C0ControlPercentEncodeSet);
          }
          this.host = result;
        } else {
          input = toASCII(input);
          if (exec(FORBIDDEN_HOST_CODE_POINT, input)) return INVALID_HOST;
          result = parseIPv4(input);
          if (result === null) return INVALID_HOST;
          this.host = result;
        }
      },
      // https://url.spec.whatwg.org/#cannot-have-a-username-password-port
      cannotHaveUsernamePasswordPort: function() {
        return !this.host || this.cannotBeABaseURL || this.scheme === "file";
      },
      // https://url.spec.whatwg.org/#include-credentials
      includesCredentials: function() {
        return this.username !== "" || this.password !== "";
      },
      // https://url.spec.whatwg.org/#is-special
      isSpecial: function() {
        return hasOwn(specialSchemes, this.scheme);
      },
      // https://url.spec.whatwg.org/#shorten-a-urls-path
      shortenPath: function() {
        var path = this.path;
        var pathSize = path.length;
        if (pathSize && (this.scheme !== "file" || pathSize !== 1 || !isWindowsDriveLetter(path[0], true))) {
          path.length--;
        }
      },
      // https://url.spec.whatwg.org/#concept-url-serializer
      serialize: function() {
        var url = this;
        var scheme = url.scheme;
        var username = url.username;
        var password = url.password;
        var host = url.host;
        var port = url.port;
        var path = url.path;
        var query = url.query;
        var fragment = url.fragment;
        var output = scheme + ":";
        if (host !== null) {
          output += "//";
          if (url.includesCredentials()) {
            output += username + (password ? ":" + password : "") + "@";
          }
          output += serializeHost(host);
          if (port !== null) output += ":" + port;
        } else if (scheme === "file") output += "//";
        output += url.cannotBeABaseURL ? path[0] : path.length ? "/" + join(path, "/") : "";
        if (query !== null) output += "?" + query;
        if (fragment !== null) output += "#" + fragment;
        return output;
      },
      // https://url.spec.whatwg.org/#dom-url-href
      setHref: function(href) {
        var failure = this.parse(href);
        if (failure) throw new TypeError2(failure);
        this.searchParams.update();
      },
      // https://url.spec.whatwg.org/#dom-url-origin
      getOrigin: function() {
        var scheme = this.scheme;
        var port = this.port;
        if (scheme === "blob") try {
          return new URLConstructor(scheme.path[0]).origin;
        } catch (error) {
          return "null";
        }
        if (scheme === "file" || !this.isSpecial()) return "null";
        return scheme + "://" + serializeHost(this.host) + (port !== null ? ":" + port : "");
      },
      // https://url.spec.whatwg.org/#dom-url-protocol
      getProtocol: function() {
        return this.scheme + ":";
      },
      setProtocol: function(protocol) {
        this.parse($toString(protocol) + ":", SCHEME_START);
      },
      // https://url.spec.whatwg.org/#dom-url-username
      getUsername: function() {
        return this.username;
      },
      setUsername: function(username) {
        var codePoints = arrayFrom($toString(username));
        if (this.cannotHaveUsernamePasswordPort()) return;
        this.username = "";
        for (var i = 0; i < codePoints.length; i++) {
          this.username += percentEncode(codePoints[i], userinfoPercentEncodeSet);
        }
      },
      // https://url.spec.whatwg.org/#dom-url-password
      getPassword: function() {
        return this.password;
      },
      setPassword: function(password) {
        var codePoints = arrayFrom($toString(password));
        if (this.cannotHaveUsernamePasswordPort()) return;
        this.password = "";
        for (var i = 0; i < codePoints.length; i++) {
          this.password += percentEncode(codePoints[i], userinfoPercentEncodeSet);
        }
      },
      // https://url.spec.whatwg.org/#dom-url-host
      getHost: function() {
        var host = this.host;
        var port = this.port;
        return host === null ? "" : port === null ? serializeHost(host) : serializeHost(host) + ":" + port;
      },
      setHost: function(host) {
        if (this.cannotBeABaseURL) return;
        this.parse(host, HOST);
      },
      // https://url.spec.whatwg.org/#dom-url-hostname
      getHostname: function() {
        var host = this.host;
        return host === null ? "" : serializeHost(host);
      },
      setHostname: function(hostname) {
        if (this.cannotBeABaseURL) return;
        this.parse(hostname, HOSTNAME);
      },
      // https://url.spec.whatwg.org/#dom-url-port
      getPort: function() {
        var port = this.port;
        return port === null ? "" : $toString(port);
      },
      setPort: function(port) {
        if (this.cannotHaveUsernamePasswordPort()) return;
        port = $toString(port);
        if (port === "") this.port = null;
        else this.parse(port, PORT);
      },
      // https://url.spec.whatwg.org/#dom-url-pathname
      getPathname: function() {
        var path = this.path;
        return this.cannotBeABaseURL ? path[0] : path.length ? "/" + join(path, "/") : "";
      },
      setPathname: function(pathname) {
        if (this.cannotBeABaseURL) return;
        this.path = [];
        this.parse(pathname, PATH_START);
      },
      // https://url.spec.whatwg.org/#dom-url-search
      getSearch: function() {
        var query = this.query;
        return query ? "?" + query : "";
      },
      setSearch: function(search) {
        search = $toString(search);
        if (search === "") {
          this.query = null;
        } else {
          if (charAt(search, 0) === "?") search = stringSlice(search, 1);
          this.query = "";
          this.parse(search, QUERY);
        }
        this.searchParams.update();
      },
      // https://url.spec.whatwg.org/#dom-url-searchparams
      getSearchParams: function() {
        return this.searchParams.facade;
      },
      // https://url.spec.whatwg.org/#dom-url-hash
      getHash: function() {
        var fragment = this.fragment;
        return fragment ? "#" + fragment : "";
      },
      setHash: function(hash) {
        hash = $toString(hash);
        if (hash === "") {
          this.fragment = null;
          return;
        }
        if (charAt(hash, 0) === "#") hash = stringSlice(hash, 1);
        this.fragment = "";
        this.parse(hash, FRAGMENT);
      },
      update: function() {
        this.query = this.searchParams.serialize() || null;
      }
    };
    var URLConstructor = function URL2(url) {
      var that = anInstance(this, URLPrototype);
      var base = validateArgumentsLength(arguments.length, 1) > 1 ? arguments[1] : void 0;
      var state = setInternalState(that, new URLState(url, false, base));
      if (!DESCRIPTORS) {
        that.href = state.serialize();
        that.origin = state.getOrigin();
        that.protocol = state.getProtocol();
        that.username = state.getUsername();
        that.password = state.getPassword();
        that.host = state.getHost();
        that.hostname = state.getHostname();
        that.port = state.getPort();
        that.pathname = state.getPathname();
        that.search = state.getSearch();
        that.searchParams = state.getSearchParams();
        that.hash = state.getHash();
      }
    };
    var URLPrototype = URLConstructor.prototype;
    var accessorDescriptor = function(getter, setter) {
      return {
        get: function() {
          return getInternalURLState(this)[getter]();
        },
        set: setter && function(value) {
          return getInternalURLState(this)[setter](value);
        },
        configurable: true,
        enumerable: true
      };
    };
    if (DESCRIPTORS) {
      defineBuiltInAccessor(URLPrototype, "href", accessorDescriptor("serialize", "setHref"));
      defineBuiltInAccessor(URLPrototype, "origin", accessorDescriptor("getOrigin"));
      defineBuiltInAccessor(URLPrototype, "protocol", accessorDescriptor("getProtocol", "setProtocol"));
      defineBuiltInAccessor(URLPrototype, "username", accessorDescriptor("getUsername", "setUsername"));
      defineBuiltInAccessor(URLPrototype, "password", accessorDescriptor("getPassword", "setPassword"));
      defineBuiltInAccessor(URLPrototype, "host", accessorDescriptor("getHost", "setHost"));
      defineBuiltInAccessor(URLPrototype, "hostname", accessorDescriptor("getHostname", "setHostname"));
      defineBuiltInAccessor(URLPrototype, "port", accessorDescriptor("getPort", "setPort"));
      defineBuiltInAccessor(URLPrototype, "pathname", accessorDescriptor("getPathname", "setPathname"));
      defineBuiltInAccessor(URLPrototype, "search", accessorDescriptor("getSearch", "setSearch"));
      defineBuiltInAccessor(URLPrototype, "searchParams", accessorDescriptor("getSearchParams"));
      defineBuiltInAccessor(URLPrototype, "hash", accessorDescriptor("getHash", "setHash"));
    }
    defineBuiltIn(URLPrototype, "toJSON", function toJSON() {
      return getInternalURLState(this).serialize();
    }, { enumerable: true });
    defineBuiltIn(URLPrototype, "toString", function toString() {
      return getInternalURLState(this).serialize();
    }, { enumerable: true });
    if (NativeURL) {
      nativeCreateObjectURL = NativeURL.createObjectURL;
      nativeRevokeObjectURL = NativeURL.revokeObjectURL;
      if (nativeCreateObjectURL) defineBuiltIn(URLConstructor, "createObjectURL", bind(nativeCreateObjectURL, NativeURL));
      if (nativeRevokeObjectURL) defineBuiltIn(URLConstructor, "revokeObjectURL", bind(nativeRevokeObjectURL, NativeURL));
    }
    var nativeCreateObjectURL;
    var nativeRevokeObjectURL;
    setToStringTag(URLConstructor, "URL");
    $({ global: true, constructor: true, forced: !USE_NATIVE_URL, sham: !DESCRIPTORS }, {
      URL: URLConstructor
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/web.url.js
var require_web_url = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/web.url.js": function() {
    "use strict";
    require_web_url_constructor();
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/web.url.to-json.js
var require_web_url_to_json = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/web.url.to-json.js": function() {
    "use strict";
    var $ = require_export();
    var call = require_function_call();
    $({ target: "URL", proto: true, enumerable: true }, {
      toJSON: function toJSON() {
        return call(URL.prototype.toString, this);
      }
    });
  }
});

// node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/web.url-search-params.js
var require_web_url_search_params = __commonJS({
  "node_modules/.pnpm/core-js@3.39.0/node_modules/core-js/modules/web.url-search-params.js": function() {
    "use strict";
    require_web_url_search_params_constructor();
  }
});

// public/assets/js/kiosk-legacy.js
require_es_symbol_async_iterator();
require_es_array_reverse();
require_es_symbol_description();
require_es_array_flat();
require_es_array_iterator();
require_es_array_unscopables_flat();
require_es_object_from_entries();
require_es_promise();
require_es_regexp_constructor();
require_es_regexp_exec();
require_es_string_replace();
require_es_string_split();
require_es_string_trim();
require_web_dom_collections_for_each();
require_web_dom_collections_iterator();
require_web_url();
require_web_url_to_json();
require_web_url_search_params();
function _regeneratorRuntime() {
  "use strict";
  _regeneratorRuntime = function _regeneratorRuntime2() {
    return e;
  };
  var t, e = {}, r = Object.prototype, n = r.hasOwnProperty, o = Object.defineProperty || function(t2, e2, r2) {
    t2[e2] = r2.value;
  }, i = "function" == typeof Symbol ? Symbol : {}, a = i.iterator || "@@iterator", c = i.asyncIterator || "@@asyncIterator", u = i.toStringTag || "@@toStringTag";
  function define(t2, e2, r2) {
    return Object.defineProperty(t2, e2, { value: r2, enumerable: true, configurable: true, writable: true }), t2[e2];
  }
  try {
    define({}, "");
  } catch (t2) {
    define = function define2(t3, e2, r2) {
      return t3[e2] = r2;
    };
  }
  function wrap(t2, e2, r2, n2) {
    var i2 = e2 && e2.prototype instanceof Generator ? e2 : Generator, a2 = Object.create(i2.prototype), c2 = new Context(n2 || []);
    return o(a2, "_invoke", { value: makeInvokeMethod(t2, r2, c2) }), a2;
  }
  function tryCatch(t2, e2, r2) {
    try {
      return { type: "normal", arg: t2.call(e2, r2) };
    } catch (t3) {
      return { type: "throw", arg: t3 };
    }
  }
  e.wrap = wrap;
  var h = "suspendedStart", l = "suspendedYield", f = "executing", s = "completed", y = {};
  function Generator() {
  }
  function GeneratorFunction() {
  }
  function GeneratorFunctionPrototype() {
  }
  var p = {};
  define(p, a, function() {
    return this;
  });
  var d = Object.getPrototypeOf, v = d && d(d(values([])));
  v && v !== r && n.call(v, a) && (p = v);
  var g = GeneratorFunctionPrototype.prototype = Generator.prototype = Object.create(p);
  function defineIteratorMethods(t2) {
    ["next", "throw", "return"].forEach(function(e2) {
      define(t2, e2, function(t3) {
        return this._invoke(e2, t3);
      });
    });
  }
  function AsyncIterator(t2, e2) {
    function invoke(r3, o2, i2, a2) {
      var c2 = tryCatch(t2[r3], t2, o2);
      if ("throw" !== c2.type) {
        var u2 = c2.arg, h2 = u2.value;
        return h2 && "object" == _typeof(h2) && n.call(h2, "__await") ? e2.resolve(h2.__await).then(function(t3) {
          invoke("next", t3, i2, a2);
        }, function(t3) {
          invoke("throw", t3, i2, a2);
        }) : e2.resolve(h2).then(function(t3) {
          u2.value = t3, i2(u2);
        }, function(t3) {
          return invoke("throw", t3, i2, a2);
        });
      }
      a2(c2.arg);
    }
    var r2;
    o(this, "_invoke", { value: function value(t3, n2) {
      function callInvokeWithMethodAndArg() {
        return new e2(function(e3, r3) {
          invoke(t3, n2, e3, r3);
        });
      }
      return r2 = r2 ? r2.then(callInvokeWithMethodAndArg, callInvokeWithMethodAndArg) : callInvokeWithMethodAndArg();
    } });
  }
  function makeInvokeMethod(e2, r2, n2) {
    var o2 = h;
    return function(i2, a2) {
      if (o2 === f) throw Error("Generator is already running");
      if (o2 === s) {
        if ("throw" === i2) throw a2;
        return { value: t, done: true };
      }
      for (n2.method = i2, n2.arg = a2; ; ) {
        var c2 = n2.delegate;
        if (c2) {
          var u2 = maybeInvokeDelegate(c2, n2);
          if (u2) {
            if (u2 === y) continue;
            return u2;
          }
        }
        if ("next" === n2.method) n2.sent = n2._sent = n2.arg;
        else if ("throw" === n2.method) {
          if (o2 === h) throw o2 = s, n2.arg;
          n2.dispatchException(n2.arg);
        } else "return" === n2.method && n2.abrupt("return", n2.arg);
        o2 = f;
        var p2 = tryCatch(e2, r2, n2);
        if ("normal" === p2.type) {
          if (o2 = n2.done ? s : l, p2.arg === y) continue;
          return { value: p2.arg, done: n2.done };
        }
        "throw" === p2.type && (o2 = s, n2.method = "throw", n2.arg = p2.arg);
      }
    };
  }
  function maybeInvokeDelegate(e2, r2) {
    var n2 = r2.method, o2 = e2.iterator[n2];
    if (o2 === t) return r2.delegate = null, "throw" === n2 && e2.iterator["return"] && (r2.method = "return", r2.arg = t, maybeInvokeDelegate(e2, r2), "throw" === r2.method) || "return" !== n2 && (r2.method = "throw", r2.arg = new TypeError("The iterator does not provide a '" + n2 + "' method")), y;
    var i2 = tryCatch(o2, e2.iterator, r2.arg);
    if ("throw" === i2.type) return r2.method = "throw", r2.arg = i2.arg, r2.delegate = null, y;
    var a2 = i2.arg;
    return a2 ? a2.done ? (r2[e2.resultName] = a2.value, r2.next = e2.nextLoc, "return" !== r2.method && (r2.method = "next", r2.arg = t), r2.delegate = null, y) : a2 : (r2.method = "throw", r2.arg = new TypeError("iterator result is not an object"), r2.delegate = null, y);
  }
  function pushTryEntry(t2) {
    var e2 = { tryLoc: t2[0] };
    1 in t2 && (e2.catchLoc = t2[1]), 2 in t2 && (e2.finallyLoc = t2[2], e2.afterLoc = t2[3]), this.tryEntries.push(e2);
  }
  function resetTryEntry(t2) {
    var e2 = t2.completion || {};
    e2.type = "normal", delete e2.arg, t2.completion = e2;
  }
  function Context(t2) {
    this.tryEntries = [{ tryLoc: "root" }], t2.forEach(pushTryEntry, this), this.reset(true);
  }
  function values(e2) {
    if (e2 || "" === e2) {
      var r2 = e2[a];
      if (r2) return r2.call(e2);
      if ("function" == typeof e2.next) return e2;
      if (!isNaN(e2.length)) {
        var o2 = -1, i2 = function next() {
          for (; ++o2 < e2.length; ) if (n.call(e2, o2)) return next.value = e2[o2], next.done = false, next;
          return next.value = t, next.done = true, next;
        };
        return i2.next = i2;
      }
    }
    throw new TypeError(_typeof(e2) + " is not iterable");
  }
  return GeneratorFunction.prototype = GeneratorFunctionPrototype, o(g, "constructor", { value: GeneratorFunctionPrototype, configurable: true }), o(GeneratorFunctionPrototype, "constructor", { value: GeneratorFunction, configurable: true }), GeneratorFunction.displayName = define(GeneratorFunctionPrototype, u, "GeneratorFunction"), e.isGeneratorFunction = function(t2) {
    var e2 = "function" == typeof t2 && t2.constructor;
    return !!e2 && (e2 === GeneratorFunction || "GeneratorFunction" === (e2.displayName || e2.name));
  }, e.mark = function(t2) {
    return Object.setPrototypeOf ? Object.setPrototypeOf(t2, GeneratorFunctionPrototype) : (t2.__proto__ = GeneratorFunctionPrototype, define(t2, u, "GeneratorFunction")), t2.prototype = Object.create(g), t2;
  }, e.awrap = function(t2) {
    return { __await: t2 };
  }, defineIteratorMethods(AsyncIterator.prototype), define(AsyncIterator.prototype, c, function() {
    return this;
  }), e.AsyncIterator = AsyncIterator, e.async = function(t2, r2, n2, o2, i2) {
    void 0 === i2 && (i2 = Promise);
    var a2 = new AsyncIterator(wrap(t2, r2, n2, o2), i2);
    return e.isGeneratorFunction(r2) ? a2 : a2.next().then(function(t3) {
      return t3.done ? t3.value : a2.next();
    });
  }, defineIteratorMethods(g), define(g, u, "Generator"), define(g, a, function() {
    return this;
  }), define(g, "toString", function() {
    return "[object Generator]";
  }), e.keys = function(t2) {
    var e2 = Object(t2), r2 = [];
    for (var n2 in e2) r2.push(n2);
    return r2.reverse(), function next() {
      for (; r2.length; ) {
        var t3 = r2.pop();
        if (t3 in e2) return next.value = t3, next.done = false, next;
      }
      return next.done = true, next;
    };
  }, e.values = values, Context.prototype = { constructor: Context, reset: function reset(e2) {
    if (this.prev = 0, this.next = 0, this.sent = this._sent = t, this.done = false, this.delegate = null, this.method = "next", this.arg = t, this.tryEntries.forEach(resetTryEntry), !e2) for (var r2 in this) "t" === r2.charAt(0) && n.call(this, r2) && !isNaN(+r2.slice(1)) && (this[r2] = t);
  }, stop: function stop() {
    this.done = true;
    var t2 = this.tryEntries[0].completion;
    if ("throw" === t2.type) throw t2.arg;
    return this.rval;
  }, dispatchException: function dispatchException(e2) {
    if (this.done) throw e2;
    var r2 = this;
    function handle(n2, o3) {
      return a2.type = "throw", a2.arg = e2, r2.next = n2, o3 && (r2.method = "next", r2.arg = t), !!o3;
    }
    for (var o2 = this.tryEntries.length - 1; o2 >= 0; --o2) {
      var i2 = this.tryEntries[o2], a2 = i2.completion;
      if ("root" === i2.tryLoc) return handle("end");
      if (i2.tryLoc <= this.prev) {
        var c2 = n.call(i2, "catchLoc"), u2 = n.call(i2, "finallyLoc");
        if (c2 && u2) {
          if (this.prev < i2.catchLoc) return handle(i2.catchLoc, true);
          if (this.prev < i2.finallyLoc) return handle(i2.finallyLoc);
        } else if (c2) {
          if (this.prev < i2.catchLoc) return handle(i2.catchLoc, true);
        } else {
          if (!u2) throw Error("try statement without catch or finally");
          if (this.prev < i2.finallyLoc) return handle(i2.finallyLoc);
        }
      }
    }
  }, abrupt: function abrupt(t2, e2) {
    for (var r2 = this.tryEntries.length - 1; r2 >= 0; --r2) {
      var o2 = this.tryEntries[r2];
      if (o2.tryLoc <= this.prev && n.call(o2, "finallyLoc") && this.prev < o2.finallyLoc) {
        var i2 = o2;
        break;
      }
    }
    i2 && ("break" === t2 || "continue" === t2) && i2.tryLoc <= e2 && e2 <= i2.finallyLoc && (i2 = null);
    var a2 = i2 ? i2.completion : {};
    return a2.type = t2, a2.arg = e2, i2 ? (this.method = "next", this.next = i2.finallyLoc, y) : this.complete(a2);
  }, complete: function complete(t2, e2) {
    if ("throw" === t2.type) throw t2.arg;
    return "break" === t2.type || "continue" === t2.type ? this.next = t2.arg : "return" === t2.type ? (this.rval = this.arg = t2.arg, this.method = "return", this.next = "end") : "normal" === t2.type && e2 && (this.next = e2), y;
  }, finish: function finish(t2) {
    for (var e2 = this.tryEntries.length - 1; e2 >= 0; --e2) {
      var r2 = this.tryEntries[e2];
      if (r2.finallyLoc === t2) return this.complete(r2.completion, r2.afterLoc), resetTryEntry(r2), y;
    }
  }, "catch": function _catch(t2) {
    for (var e2 = this.tryEntries.length - 1; e2 >= 0; --e2) {
      var r2 = this.tryEntries[e2];
      if (r2.tryLoc === t2) {
        var n2 = r2.completion;
        if ("throw" === n2.type) {
          var o2 = n2.arg;
          resetTryEntry(r2);
        }
        return o2;
      }
    }
    throw Error("illegal catch attempt");
  }, delegateYield: function delegateYield(e2, r2, n2) {
    return this.delegate = { iterator: values(e2), resultName: r2, nextLoc: n2 }, "next" === this.method && (this.arg = t), y;
  } }, e;
}
function _slicedToArray(r, e) {
  return _arrayWithHoles(r) || _iterableToArrayLimit(r, e) || _unsupportedIterableToArray(r, e) || _nonIterableRest();
}
function _nonIterableRest() {
  throw new TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
}
function _iterableToArrayLimit(r, l) {
  var t = null == r ? null : "undefined" != typeof Symbol && r[Symbol.iterator] || r["@@iterator"];
  if (null != t) {
    var e, n, i, u, a = [], f = true, o = false;
    try {
      if (i = (t = t.call(r)).next, 0 === l) {
        if (Object(t) !== t) return;
        f = false;
      } else for (; !(f = (e = i.call(t)).done) && (a.push(e.value), a.length !== l); f = true) ;
    } catch (r2) {
      o = true, n = r2;
    } finally {
      try {
        if (!f && null != t["return"] && (u = t["return"](), Object(u) !== u)) return;
      } finally {
        if (o) throw n;
      }
    }
    return a;
  }
}
function _arrayWithHoles(r) {
  if (Array.isArray(r)) return r;
}
function _toConsumableArray(r) {
  return _arrayWithoutHoles(r) || _iterableToArray(r) || _unsupportedIterableToArray(r) || _nonIterableSpread();
}
function _nonIterableSpread() {
  throw new TypeError("Invalid attempt to spread non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
}
function _iterableToArray(r) {
  if ("undefined" != typeof Symbol && null != r[Symbol.iterator] || null != r["@@iterator"]) return Array.from(r);
}
function _arrayWithoutHoles(r) {
  if (Array.isArray(r)) return _arrayLikeToArray(r);
}
function _createForOfIteratorHelper(r, e) {
  var t = "undefined" != typeof Symbol && r[Symbol.iterator] || r["@@iterator"];
  if (!t) {
    if (Array.isArray(r) || (t = _unsupportedIterableToArray(r)) || e && r && "number" == typeof r.length) {
      t && (r = t);
      var _n = 0, F = function F2() {
      };
      return { s: F, n: function n() {
        return _n >= r.length ? { done: true } : { done: false, value: r[_n++] };
      }, e: function e2(r2) {
        throw r2;
      }, f: F };
    }
    throw new TypeError("Invalid attempt to iterate non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
  }
  var o, a = true, u = false;
  return { s: function s() {
    t = t.call(r);
  }, n: function n() {
    var r2 = t.next();
    return a = r2.done, r2;
  }, e: function e2(r2) {
    u = true, o = r2;
  }, f: function f() {
    try {
      a || null == t["return"] || t["return"]();
    } finally {
      if (u) throw o;
    }
  } };
}
function _unsupportedIterableToArray(r, a) {
  if (r) {
    if ("string" == typeof r) return _arrayLikeToArray(r, a);
    var t = {}.toString.call(r).slice(8, -1);
    return "Object" === t && r.constructor && (t = r.constructor.name), "Map" === t || "Set" === t ? Array.from(r) : "Arguments" === t || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(t) ? _arrayLikeToArray(r, a) : void 0;
  }
}
function _arrayLikeToArray(r, a) {
  (null == a || a > r.length) && (a = r.length);
  for (var e = 0, n = Array(a); e < a; e++) n[e] = r[e];
  return n;
}
function _typeof(o) {
  "@babel/helpers - typeof";
  return _typeof = "function" == typeof Symbol && "symbol" == typeof Symbol.iterator ? function(o2) {
    return typeof o2;
  } : function(o2) {
    return o2 && "function" == typeof Symbol && o2.constructor === Symbol && o2 !== Symbol.prototype ? "symbol" : typeof o2;
  }, _typeof(o);
}
var kiosk = function() {
  var __defProp = Object.defineProperty;
  var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
  var __getOwnPropNames = Object.getOwnPropertyNames;
  var __hasOwnProp = Object.prototype.hasOwnProperty;
  var __export = function __export2(target, all) {
    for (var name in all) __defProp(target, name, {
      get: all[name],
      enumerable: true
    });
  };
  var __copyProps = function __copyProps2(to, from, except, desc) {
    if (from && _typeof(from) === "object" || typeof from === "function") {
      var _iterator = _createForOfIteratorHelper(__getOwnPropNames(from)), _step;
      try {
        var _loop = function _loop2() {
          var key = _step.value;
          if (!__hasOwnProp.call(to, key) && key !== except) __defProp(to, key, {
            get: function get() {
              return from[key];
            },
            enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable
          });
        };
        for (_iterator.s(); !(_step = _iterator.n()).done; ) {
          _loop();
        }
      } catch (err) {
        _iterator.e(err);
      } finally {
        _iterator.f();
      }
    }
    return to;
  };
  var __toCommonJS = function __toCommonJS2(mod) {
    return __copyProps(__defProp({}, "__esModule", {
      value: true
    }), mod);
  };
  var __async = function __async2(__this, __arguments, generator) {
    return new Promise(function(resolve, reject) {
      var fulfilled = function fulfilled2(value) {
        try {
          step(generator.next(value));
        } catch (e) {
          reject(e);
        }
      };
      var rejected = function rejected2(value) {
        try {
          step(generator["throw"](value));
        } catch (e) {
          reject(e);
        }
      };
      var step = function step2(x) {
        return x.done ? resolve(x.value) : Promise.resolve(x.value).then(fulfilled, rejected);
      };
      step((generator = generator.apply(__this, __arguments)).next());
    });
  };
  var kiosk_exports = {};
  __export(kiosk_exports, {
    cleanupFrames: function cleanupFrames() {
      return _cleanupFrames;
    },
    startPolling: function startPolling() {
      return _startPolling;
    }
  });
  var htmx2 = function() {
    "use strict";
    var htmx = {
      // Tsc madness here, assigning the functions directly results in an invalid TypeScript output, but reassigning is fine
      /* Event processing */
      /** @type {typeof onLoadHelper} */
      onLoad: null,
      /** @type {typeof processNode} */
      process: null,
      /** @type {typeof addEventListenerImpl} */
      on: null,
      /** @type {typeof removeEventListenerImpl} */
      off: null,
      /** @type {typeof triggerEvent} */
      trigger: null,
      /** @type {typeof ajaxHelper} */
      ajax: null,
      /* DOM querying helpers */
      /** @type {typeof find} */
      find: null,
      /** @type {typeof findAll} */
      findAll: null,
      /** @type {typeof closest} */
      closest: null,
      /**
       * Returns the input values that would resolve for a given element via the htmx value resolution mechanism
       *
       * @see https://htmx.org/api/#values
       *
       * @param {Element} elt the element to resolve values on
       * @param {HttpVerb} type the request type (e.g. **get** or **post**) non-GET's will include the enclosing form of the element. Defaults to **post**
       * @returns {Object}
       */
      values: function values(elt, type) {
        var inputValues = getInputValues(elt, type || "post");
        return inputValues.values;
      },
      /* DOM manipulation helpers */
      /** @type {typeof removeElement} */
      remove: null,
      /** @type {typeof addClassToElement} */
      addClass: null,
      /** @type {typeof removeClassFromElement} */
      removeClass: null,
      /** @type {typeof toggleClassOnElement} */
      toggleClass: null,
      /** @type {typeof takeClassForElement} */
      takeClass: null,
      /** @type {typeof swap} */
      swap: null,
      /* Extension entrypoints */
      /** @type {typeof defineExtension} */
      defineExtension: null,
      /** @type {typeof removeExtension} */
      removeExtension: null,
      /* Debugging */
      /** @type {typeof logAll} */
      logAll: null,
      /** @type {typeof logNone} */
      logNone: null,
      /* Debugging */
      /**
       * The logger htmx uses to log with
       *
       * @see https://htmx.org/api/#logger
       */
      logger: null,
      /**
       * A property holding the configuration htmx uses at runtime.
       *
       * Note that using a [meta tag](https://htmx.org/docs/#config) is the preferred mechanism for setting these properties.
       *
       * @see https://htmx.org/api/#config
       */
      config: {
        /**
         * Whether to use history.
         * @type boolean
         * @default true
         */
        historyEnabled: true,
        /**
         * The number of pages to keep in **localStorage** for history support.
         * @type number
         * @default 10
         */
        historyCacheSize: 10,
        /**
         * @type boolean
         * @default false
         */
        refreshOnHistoryMiss: false,
        /**
         * The default swap style to use if **[hx-swap](https://htmx.org/attributes/hx-swap)** is omitted.
         * @type HtmxSwapStyle
         * @default 'innerHTML'
         */
        defaultSwapStyle: "innerHTML",
        /**
         * The default delay between receiving a response from the server and doing the swap.
         * @type number
         * @default 0
         */
        defaultSwapDelay: 0,
        /**
         * The default delay between completing the content swap and settling attributes.
         * @type number
         * @default 20
         */
        defaultSettleDelay: 20,
        /**
         * If true, htmx will inject a small amount of CSS into the page to make indicators invisible unless the **htmx-indicator** class is present.
         * @type boolean
         * @default true
         */
        includeIndicatorStyles: true,
        /**
         * The class to place on indicators when a request is in flight.
         * @type string
         * @default 'htmx-indicator'
         */
        indicatorClass: "htmx-indicator",
        /**
         * The class to place on triggering elements when a request is in flight.
         * @type string
         * @default 'htmx-request'
         */
        requestClass: "htmx-request",
        /**
         * The class to temporarily place on elements that htmx has added to the DOM.
         * @type string
         * @default 'htmx-added'
         */
        addedClass: "htmx-added",
        /**
         * The class to place on target elements when htmx is in the settling phase.
         * @type string
         * @default 'htmx-settling'
         */
        settlingClass: "htmx-settling",
        /**
         * The class to place on target elements when htmx is in the swapping phase.
         * @type string
         * @default 'htmx-swapping'
         */
        swappingClass: "htmx-swapping",
        /**
         * Allows the use of eval-like functionality in htmx, to enable **hx-vars**, trigger conditions & script tag evaluation. Can be set to **false** for CSP compatibility.
         * @type boolean
         * @default true
         */
        allowEval: true,
        /**
         * If set to false, disables the interpretation of script tags.
         * @type boolean
         * @default true
         */
        allowScriptTags: true,
        /**
         * If set, the nonce will be added to inline scripts.
         * @type string
         * @default ''
         */
        inlineScriptNonce: "",
        /**
         * If set, the nonce will be added to inline styles.
         * @type string
         * @default ''
         */
        inlineStyleNonce: "",
        /**
         * The attributes to settle during the settling phase.
         * @type string[]
         * @default ['class', 'style', 'width', 'height']
         */
        attributesToSettle: ["class", "style", "width", "height"],
        /**
         * Allow cross-site Access-Control requests using credentials such as cookies, authorization headers or TLS client certificates.
         * @type boolean
         * @default false
         */
        withCredentials: false,
        /**
         * @type number
         * @default 0
         */
        timeout: 0,
        /**
         * The default implementation of **getWebSocketReconnectDelay** for reconnecting after unexpected connection loss by the event code **Abnormal Closure**, **Service Restart** or **Try Again Later**.
         * @type {'full-jitter' | ((retryCount:number) => number)}
         * @default "full-jitter"
         */
        wsReconnectDelay: "full-jitter",
        /**
         * The type of binary data being received over the WebSocket connection
         * @type BinaryType
         * @default 'blob'
         */
        wsBinaryType: "blob",
        /**
         * @type string
         * @default '[hx-disable], [data-hx-disable]'
         */
        disableSelector: "[hx-disable], [data-hx-disable]",
        /**
         * @type {'auto' | 'instant' | 'smooth'}
         * @default 'instant'
         */
        scrollBehavior: "instant",
        /**
         * If the focused element should be scrolled into view.
         * @type boolean
         * @default false
         */
        defaultFocusScroll: false,
        /**
         * If set to true htmx will include a cache-busting parameter in GET requests to avoid caching partial responses by the browser
         * @type boolean
         * @default false
         */
        getCacheBusterParam: false,
        /**
         * If set to true, htmx will use the View Transition API when swapping in new content.
         * @type boolean
         * @default false
         */
        globalViewTransitions: false,
        /**
         * htmx will format requests with these methods by encoding their parameters in the URL, not the request body
         * @type {(HttpVerb)[]}
         * @default ['get', 'delete']
         */
        methodsThatUseUrlParams: ["get", "delete"],
        /**
         * If set to true, disables htmx-based requests to non-origin hosts.
         * @type boolean
         * @default false
         */
        selfRequestsOnly: true,
        /**
         * If set to true htmx will not update the title of the document when a title tag is found in new content
         * @type boolean
         * @default false
         */
        ignoreTitle: false,
        /**
         * Whether the target of a boosted element is scrolled into the viewport.
         * @type boolean
         * @default true
         */
        scrollIntoViewOnBoost: true,
        /**
         * The cache to store evaluated trigger specifications into.
         * You may define a simple object to use a never-clearing cache, or implement your own system using a [proxy object](https://developer.mozilla.org/docs/Web/JavaScript/Reference/Global_Objects/Proxy)
         * @type {Object|null}
         * @default null
         */
        triggerSpecsCache: null,
        /** @type boolean */
        disableInheritance: false,
        /** @type HtmxResponseHandlingConfig[] */
        responseHandling: [{
          code: "204",
          swap: false
        }, {
          code: "[23]..",
          swap: true
        }, {
          code: "[45]..",
          swap: false,
          error: true
        }],
        /**
         * Whether to process OOB swaps on elements that are nested within the main response element.
         * @type boolean
         * @default true
         */
        allowNestedOobSwaps: true
      },
      /** @type {typeof parseInterval} */
      parseInterval: null,
      /** @type {typeof internalEval} */
      _: null,
      version: "2.0.3"
    };
    htmx.onLoad = onLoadHelper;
    htmx.process = processNode;
    htmx.on = addEventListenerImpl;
    htmx.off = removeEventListenerImpl;
    htmx.trigger = triggerEvent;
    htmx.ajax = ajaxHelper;
    htmx.find = find;
    htmx.findAll = findAll;
    htmx.closest = closest;
    htmx.remove = removeElement;
    htmx.addClass = addClassToElement;
    htmx.removeClass = removeClassFromElement;
    htmx.toggleClass = toggleClassOnElement;
    htmx.takeClass = takeClassForElement;
    htmx.swap = swap;
    htmx.defineExtension = defineExtension;
    htmx.removeExtension = removeExtension;
    htmx.logAll = logAll;
    htmx.logNone = logNone;
    htmx.parseInterval = parseInterval;
    htmx._ = internalEval;
    var internalAPI = {
      addTriggerHandler: addTriggerHandler,
      bodyContains: bodyContains,
      canAccessLocalStorage: canAccessLocalStorage,
      findThisElement: findThisElement,
      filterValues: filterValues,
      swap: swap,
      hasAttribute: hasAttribute,
      getAttributeValue: getAttributeValue,
      getClosestAttributeValue: getClosestAttributeValue,
      getClosestMatch: getClosestMatch,
      getExpressionVars: getExpressionVars,
      getHeaders: getHeaders,
      getInputValues: getInputValues,
      getInternalData: getInternalData,
      getSwapSpecification: getSwapSpecification,
      getTriggerSpecs: getTriggerSpecs,
      getTarget: getTarget,
      makeFragment: makeFragment,
      mergeObjects: mergeObjects,
      makeSettleInfo: makeSettleInfo,
      oobSwap: oobSwap,
      querySelectorExt: querySelectorExt,
      settleImmediately: settleImmediately,
      shouldCancel: shouldCancel,
      triggerEvent: triggerEvent,
      triggerErrorEvent: triggerErrorEvent,
      withExtensions: withExtensions
    };
    var VERBS = ["get", "post", "put", "delete", "patch"];
    var VERB_SELECTOR = VERBS.map(function(verb) {
      return "[hx-" + verb + "], [data-hx-" + verb + "]";
    }).join(", ");
    function parseInterval(str2) {
      if (str2 == void 0) {
        return void 0;
      }
      var interval = NaN;
      if (str2.slice(-2) == "ms") {
        interval = parseFloat(str2.slice(0, -2));
      } else if (str2.slice(-1) == "s") {
        interval = parseFloat(str2.slice(0, -1)) * 1e3;
      } else if (str2.slice(-1) == "m") {
        interval = parseFloat(str2.slice(0, -1)) * 1e3 * 60;
      } else {
        interval = parseFloat(str2);
      }
      return isNaN(interval) ? void 0 : interval;
    }
    function getRawAttribute(elt, name) {
      return elt instanceof Element && elt.getAttribute(name);
    }
    function hasAttribute(elt, qualifiedName) {
      return !!elt.hasAttribute && (elt.hasAttribute(qualifiedName) || elt.hasAttribute("data-" + qualifiedName));
    }
    function getAttributeValue(elt, qualifiedName) {
      return getRawAttribute(elt, qualifiedName) || getRawAttribute(elt, "data-" + qualifiedName);
    }
    function parentElt(elt) {
      var parent = elt.parentElement;
      if (!parent && elt.parentNode instanceof ShadowRoot) return elt.parentNode;
      return parent;
    }
    function getDocument() {
      return document;
    }
    function getRootNode(elt, global2) {
      return elt.getRootNode ? elt.getRootNode({
        composed: global2
      }) : getDocument();
    }
    function getClosestMatch(elt, condition) {
      while (elt && !condition(elt)) {
        elt = parentElt(elt);
      }
      return elt || null;
    }
    function getAttributeValueWithDisinheritance(initialElement, ancestor, attributeName) {
      var attributeValue = getAttributeValue(ancestor, attributeName);
      var disinherit = getAttributeValue(ancestor, "hx-disinherit");
      var inherit = getAttributeValue(ancestor, "hx-inherit");
      if (initialElement !== ancestor) {
        if (htmx.config.disableInheritance) {
          if (inherit && (inherit === "*" || inherit.split(" ").indexOf(attributeName) >= 0)) {
            return attributeValue;
          } else {
            return null;
          }
        }
        if (disinherit && (disinherit === "*" || disinherit.split(" ").indexOf(attributeName) >= 0)) {
          return "unset";
        }
      }
      return attributeValue;
    }
    function getClosestAttributeValue(elt, attributeName) {
      var closestAttr = null;
      getClosestMatch(elt, function(e) {
        return !!(closestAttr = getAttributeValueWithDisinheritance(elt, asElement(e), attributeName));
      });
      if (closestAttr !== "unset") {
        return closestAttr;
      }
    }
    function matches(elt, selector) {
      var matchesFunction = elt instanceof Element && (elt.matches || elt.matchesSelector || elt.msMatchesSelector || elt.mozMatchesSelector || elt.webkitMatchesSelector || elt.oMatchesSelector);
      return !!matchesFunction && matchesFunction.call(elt, selector);
    }
    function getStartTag(str2) {
      var tagMatcher = /<([a-z][^\/\0>\x20\t\r\n\f]*)/i;
      var match = tagMatcher.exec(str2);
      if (match) {
        return match[1].toLowerCase();
      } else {
        return "";
      }
    }
    function parseHTML(resp) {
      var parser = new DOMParser();
      return parser.parseFromString(resp, "text/html");
    }
    function takeChildrenFor(fragment, elt) {
      while (elt.childNodes.length > 0) {
        fragment.append(elt.childNodes[0]);
      }
    }
    function duplicateScript(script) {
      var newScript = getDocument().createElement("script");
      forEach(script.attributes, function(attr) {
        newScript.setAttribute(attr.name, attr.value);
      });
      newScript.textContent = script.textContent;
      newScript.async = false;
      if (htmx.config.inlineScriptNonce) {
        newScript.nonce = htmx.config.inlineScriptNonce;
      }
      return newScript;
    }
    function isJavaScriptScriptNode(script) {
      return script.matches("script") && (script.type === "text/javascript" || script.type === "module" || script.type === "");
    }
    function normalizeScriptTags(fragment) {
      Array.from(fragment.querySelectorAll("script")).forEach(
        /** @param {HTMLScriptElement} script */
        function(script) {
          if (isJavaScriptScriptNode(script)) {
            var newScript = duplicateScript(script);
            var parent = script.parentNode;
            try {
              parent.insertBefore(newScript, script);
            } catch (e) {
              logError(e);
            } finally {
              script.remove();
            }
          }
        }
      );
    }
    function makeFragment(response) {
      var responseWithNoHead = response.replace(/<head(\s[^>]*)?>[\s\S]*?<\/head>/i, "");
      var startTag = getStartTag(responseWithNoHead);
      var fragment;
      if (startTag === "html") {
        fragment = /** @type DocumentFragmentWithTitle */
        new DocumentFragment();
        var doc = parseHTML(response);
        takeChildrenFor(fragment, doc.body);
        fragment.title = doc.title;
      } else if (startTag === "body") {
        fragment = /** @type DocumentFragmentWithTitle */
        new DocumentFragment();
        var _doc = parseHTML(responseWithNoHead);
        takeChildrenFor(fragment, _doc.body);
        fragment.title = _doc.title;
      } else {
        var _doc2 = parseHTML('<body><template class="internal-htmx-wrapper">' + responseWithNoHead + "</template></body>");
        fragment = /** @type DocumentFragmentWithTitle */
        _doc2.querySelector("template").content;
        fragment.title = _doc2.title;
        var titleElement = fragment.querySelector("title");
        if (titleElement && titleElement.parentNode === fragment) {
          titleElement.remove();
          fragment.title = titleElement.innerText;
        }
      }
      if (fragment) {
        if (htmx.config.allowScriptTags) {
          normalizeScriptTags(fragment);
        } else {
          fragment.querySelectorAll("script").forEach(function(script) {
            return script.remove();
          });
        }
      }
      return fragment;
    }
    function maybeCall(func) {
      if (func) {
        func();
      }
    }
    function isType(o, type) {
      return Object.prototype.toString.call(o) === "[object " + type + "]";
    }
    function isFunction(o) {
      return typeof o === "function";
    }
    function isRawObject(o) {
      return isType(o, "Object");
    }
    function getInternalData(elt) {
      var dataProp = "htmx-internal-data";
      var data = elt[dataProp];
      if (!data) {
        data = elt[dataProp] = {};
      }
      return data;
    }
    function toArray(arr) {
      var returnArr = [];
      if (arr) {
        for (var i = 0; i < arr.length; i++) {
          returnArr.push(arr[i]);
        }
      }
      return returnArr;
    }
    function forEach(arr, func) {
      if (arr) {
        for (var i = 0; i < arr.length; i++) {
          func(arr[i]);
        }
      }
    }
    function isScrolledIntoView(el) {
      var rect = el.getBoundingClientRect();
      var elemTop = rect.top;
      var elemBottom = rect.bottom;
      return elemTop < window.innerHeight && elemBottom >= 0;
    }
    function bodyContains(elt) {
      var rootNode = elt.getRootNode && elt.getRootNode();
      if (rootNode && rootNode instanceof window.ShadowRoot) {
        return getDocument().body.contains(rootNode.host);
      } else {
        return getDocument().body.contains(elt);
      }
    }
    function splitOnWhitespace(trigger) {
      return trigger.trim().split(/\s+/);
    }
    function mergeObjects(obj1, obj2) {
      for (var key in obj2) {
        if (obj2.hasOwnProperty(key)) {
          obj1[key] = obj2[key];
        }
      }
      return obj1;
    }
    function parseJSON(jString) {
      try {
        return JSON.parse(jString);
      } catch (error) {
        logError(error);
        return null;
      }
    }
    function canAccessLocalStorage() {
      var test = "htmx:localStorageTest";
      try {
        localStorage.setItem(test, test);
        localStorage.removeItem(test);
        return true;
      } catch (e) {
        return false;
      }
    }
    function normalizePath(path) {
      try {
        var url = new URL(path);
        if (url) {
          path = url.pathname + url.search;
        }
        if (!/^\/$/.test(path)) {
          path = path.replace(/\/+$/, "");
        }
        return path;
      } catch (e) {
        return path;
      }
    }
    function internalEval(str) {
      return maybeEval(getDocument().body, function() {
        return eval(str);
      });
    }
    function onLoadHelper(callback) {
      var value = htmx.on(
        "htmx:load",
        /** @param {CustomEvent} evt */
        function(evt) {
          callback(evt.detail.elt);
        }
      );
      return value;
    }
    function logAll() {
      htmx.logger = function(elt, event2, data) {
        if (console) {
          console.log(event2, elt, data);
        }
      };
    }
    function logNone() {
      htmx.logger = null;
    }
    function find(eltOrSelector, selector) {
      if (typeof eltOrSelector !== "string") {
        return eltOrSelector.querySelector(selector);
      } else {
        return find(getDocument(), eltOrSelector);
      }
    }
    function findAll(eltOrSelector, selector) {
      if (typeof eltOrSelector !== "string") {
        return eltOrSelector.querySelectorAll(selector);
      } else {
        return findAll(getDocument(), eltOrSelector);
      }
    }
    function getWindow() {
      return window;
    }
    function removeElement(elt, delay) {
      elt = resolveTarget(elt);
      if (delay) {
        getWindow().setTimeout(function() {
          removeElement(elt);
          elt = null;
        }, delay);
      } else {
        parentElt(elt).removeChild(elt);
      }
    }
    function asElement(elt) {
      return elt instanceof Element ? elt : null;
    }
    function asHtmlElement(elt) {
      return elt instanceof HTMLElement ? elt : null;
    }
    function asString(value) {
      return typeof value === "string" ? value : null;
    }
    function asParentNode(elt) {
      return elt instanceof Element || elt instanceof Document || elt instanceof DocumentFragment ? elt : null;
    }
    function addClassToElement(elt, clazz, delay) {
      elt = asElement(resolveTarget(elt));
      if (!elt) {
        return;
      }
      if (delay) {
        getWindow().setTimeout(function() {
          addClassToElement(elt, clazz);
          elt = null;
        }, delay);
      } else {
        elt.classList && elt.classList.add(clazz);
      }
    }
    function removeClassFromElement(node, clazz, delay) {
      var elt = asElement(resolveTarget(node));
      if (!elt) {
        return;
      }
      if (delay) {
        getWindow().setTimeout(function() {
          removeClassFromElement(elt, clazz);
          elt = null;
        }, delay);
      } else {
        if (elt.classList) {
          elt.classList.remove(clazz);
          if (elt.classList.length === 0) {
            elt.removeAttribute("class");
          }
        }
      }
    }
    function toggleClassOnElement(elt, clazz) {
      elt = resolveTarget(elt);
      elt.classList.toggle(clazz);
    }
    function takeClassForElement(elt, clazz) {
      elt = resolveTarget(elt);
      forEach(elt.parentElement.children, function(child) {
        removeClassFromElement(child, clazz);
      });
      addClassToElement(asElement(elt), clazz);
    }
    function closest(elt, selector) {
      elt = asElement(resolveTarget(elt));
      if (elt && elt.closest) {
        return elt.closest(selector);
      } else {
        do {
          if (elt == null || matches(elt, selector)) {
            return elt;
          }
        } while (elt = elt && asElement(parentElt(elt)));
        return null;
      }
    }
    function startsWith(str2, prefix) {
      return str2.substring(0, prefix.length) === prefix;
    }
    function endsWith(str2, suffix) {
      return str2.substring(str2.length - suffix.length) === suffix;
    }
    function normalizeSelector(selector) {
      var trimmedSelector = selector.trim();
      if (startsWith(trimmedSelector, "<") && endsWith(trimmedSelector, "/>")) {
        return trimmedSelector.substring(1, trimmedSelector.length - 2);
      } else {
        return trimmedSelector;
      }
    }
    function querySelectorAllExt(elt, selector, global2) {
      elt = resolveTarget(elt);
      if (selector.indexOf("closest ") === 0) {
        return [closest(asElement(elt), normalizeSelector(selector.substr(8)))];
      } else if (selector.indexOf("find ") === 0) {
        return [find(asParentNode(elt), normalizeSelector(selector.substr(5)))];
      } else if (selector === "next") {
        return [asElement(elt).nextElementSibling];
      } else if (selector.indexOf("next ") === 0) {
        return [scanForwardQuery(elt, normalizeSelector(selector.substr(5)), !!global2)];
      } else if (selector === "previous") {
        return [asElement(elt).previousElementSibling];
      } else if (selector.indexOf("previous ") === 0) {
        return [scanBackwardsQuery(elt, normalizeSelector(selector.substr(9)), !!global2)];
      } else if (selector === "document") {
        return [document];
      } else if (selector === "window") {
        return [window];
      } else if (selector === "body") {
        return [document.body];
      } else if (selector === "root") {
        return [getRootNode(elt, !!global2)];
      } else if (selector === "host") {
        return [
          /** @type ShadowRoot */
          elt.getRootNode().host
        ];
      } else if (selector.indexOf("global ") === 0) {
        return querySelectorAllExt(elt, selector.slice(7), true);
      } else {
        return toArray(asParentNode(getRootNode(elt, !!global2)).querySelectorAll(normalizeSelector(selector)));
      }
    }
    var scanForwardQuery = function scanForwardQuery2(start, match, global2) {
      var results = asParentNode(getRootNode(start, global2)).querySelectorAll(match);
      for (var i = 0; i < results.length; i++) {
        var elt = results[i];
        if (elt.compareDocumentPosition(start) === Node.DOCUMENT_POSITION_PRECEDING) {
          return elt;
        }
      }
    };
    var scanBackwardsQuery = function scanBackwardsQuery2(start, match, global2) {
      var results = asParentNode(getRootNode(start, global2)).querySelectorAll(match);
      for (var i = results.length - 1; i >= 0; i--) {
        var elt = results[i];
        if (elt.compareDocumentPosition(start) === Node.DOCUMENT_POSITION_FOLLOWING) {
          return elt;
        }
      }
    };
    function querySelectorExt(eltOrSelector, selector) {
      if (typeof eltOrSelector !== "string") {
        return querySelectorAllExt(eltOrSelector, selector)[0];
      } else {
        return querySelectorAllExt(getDocument().body, eltOrSelector)[0];
      }
    }
    function resolveTarget(eltOrSelector, context) {
      if (typeof eltOrSelector === "string") {
        return find(asParentNode(context) || document, eltOrSelector);
      } else {
        return eltOrSelector;
      }
    }
    function processEventArgs(arg1, arg2, arg3, arg4) {
      if (isFunction(arg2)) {
        return {
          target: getDocument().body,
          event: asString(arg1),
          listener: arg2,
          options: arg3
        };
      } else {
        return {
          target: resolveTarget(arg1),
          event: asString(arg2),
          listener: arg3,
          options: arg4
        };
      }
    }
    function addEventListenerImpl(arg1, arg2, arg3, arg4) {
      ready(function() {
        var eventArgs = processEventArgs(arg1, arg2, arg3, arg4);
        eventArgs.target.addEventListener(eventArgs.event, eventArgs.listener, eventArgs.options);
      });
      var b = isFunction(arg2);
      return b ? arg2 : arg3;
    }
    function removeEventListenerImpl(arg1, arg2, arg3) {
      ready(function() {
        var eventArgs = processEventArgs(arg1, arg2, arg3);
        eventArgs.target.removeEventListener(eventArgs.event, eventArgs.listener);
      });
      return isFunction(arg2) ? arg2 : arg3;
    }
    var DUMMY_ELT = getDocument().createElement("output");
    function findAttributeTargets(elt, attrName) {
      var attrTarget = getClosestAttributeValue(elt, attrName);
      if (attrTarget) {
        if (attrTarget === "this") {
          return [findThisElement(elt, attrName)];
        } else {
          var result = querySelectorAllExt(elt, attrTarget);
          if (result.length === 0) {
            logError('The selector "' + attrTarget + '" on ' + attrName + " returned no matches!");
            return [DUMMY_ELT];
          } else {
            return result;
          }
        }
      }
    }
    function findThisElement(elt, attribute) {
      return asElement(getClosestMatch(elt, function(elt2) {
        return getAttributeValue(asElement(elt2), attribute) != null;
      }));
    }
    function getTarget(elt) {
      var targetStr = getClosestAttributeValue(elt, "hx-target");
      if (targetStr) {
        if (targetStr === "this") {
          return findThisElement(elt, "hx-target");
        } else {
          return querySelectorExt(elt, targetStr);
        }
      } else {
        var data = getInternalData(elt);
        if (data.boosted) {
          return getDocument().body;
        } else {
          return elt;
        }
      }
    }
    function shouldSettleAttribute(name) {
      var attributesToSettle = htmx.config.attributesToSettle;
      for (var i = 0; i < attributesToSettle.length; i++) {
        if (name === attributesToSettle[i]) {
          return true;
        }
      }
      return false;
    }
    function cloneAttributes(mergeTo, mergeFrom) {
      forEach(mergeTo.attributes, function(attr) {
        if (!mergeFrom.hasAttribute(attr.name) && shouldSettleAttribute(attr.name)) {
          mergeTo.removeAttribute(attr.name);
        }
      });
      forEach(mergeFrom.attributes, function(attr) {
        if (shouldSettleAttribute(attr.name)) {
          mergeTo.setAttribute(attr.name, attr.value);
        }
      });
    }
    function isInlineSwap(swapStyle, target) {
      var extensions2 = getExtensions(target);
      for (var i = 0; i < extensions2.length; i++) {
        var extension = extensions2[i];
        try {
          if (extension.isInlineSwap(swapStyle)) {
            return true;
          }
        } catch (e) {
          logError(e);
        }
      }
      return swapStyle === "outerHTML";
    }
    function oobSwap(oobValue, oobElement, settleInfo, rootNode) {
      rootNode = rootNode || getDocument();
      var selector = "#" + getRawAttribute(oobElement, "id");
      var swapStyle = "outerHTML";
      if (oobValue === "true") {
      } else if (oobValue.indexOf(":") > 0) {
        swapStyle = oobValue.substr(0, oobValue.indexOf(":"));
        selector = oobValue.substr(oobValue.indexOf(":") + 1, oobValue.length);
      } else {
        swapStyle = oobValue;
      }
      oobElement.removeAttribute("hx-swap-oob");
      oobElement.removeAttribute("data-hx-swap-oob");
      var targets = querySelectorAllExt(rootNode, selector, false);
      if (targets) {
        forEach(targets, function(target) {
          var fragment;
          var oobElementClone = oobElement.cloneNode(true);
          fragment = getDocument().createDocumentFragment();
          fragment.appendChild(oobElementClone);
          if (!isInlineSwap(swapStyle, target)) {
            fragment = asParentNode(oobElementClone);
          }
          var beforeSwapDetails = {
            shouldSwap: true,
            target: target,
            fragment: fragment
          };
          if (!triggerEvent(target, "htmx:oobBeforeSwap", beforeSwapDetails)) return;
          target = beforeSwapDetails.target;
          if (beforeSwapDetails.shouldSwap) {
            handlePreservedElements(fragment);
            swapWithStyle(swapStyle, target, target, fragment, settleInfo);
            restorePreservedElements();
          }
          forEach(settleInfo.elts, function(elt) {
            triggerEvent(elt, "htmx:oobAfterSwap", beforeSwapDetails);
          });
        });
        oobElement.parentNode.removeChild(oobElement);
      } else {
        oobElement.parentNode.removeChild(oobElement);
        triggerErrorEvent(getDocument().body, "htmx:oobErrorNoTarget", {
          content: oobElement
        });
      }
      return oobValue;
    }
    function restorePreservedElements() {
      var pantry = find("#--htmx-preserve-pantry--");
      if (pantry) {
        for (var _i = 0, _arr = _toConsumableArray(pantry.children); _i < _arr.length; _i++) {
          var preservedElt = _arr[_i];
          var existingElement = find("#" + preservedElt.id);
          existingElement.parentNode.moveBefore(preservedElt, existingElement);
          existingElement.remove();
        }
        pantry.remove();
      }
    }
    function handlePreservedElements(fragment) {
      forEach(findAll(fragment, "[hx-preserve], [data-hx-preserve]"), function(preservedElt) {
        var id = getAttributeValue(preservedElt, "id");
        var existingElement = getDocument().getElementById(id);
        if (existingElement != null) {
          if (preservedElt.moveBefore) {
            var pantry = find("#--htmx-preserve-pantry--");
            if (pantry == null) {
              getDocument().body.insertAdjacentHTML("afterend", "<div id='--htmx-preserve-pantry--'></div>");
              pantry = find("#--htmx-preserve-pantry--");
            }
            pantry.moveBefore(existingElement, null);
          } else {
            preservedElt.parentNode.replaceChild(existingElement, preservedElt);
          }
        }
      });
    }
    function handleAttributes(parentNode, fragment, settleInfo) {
      forEach(fragment.querySelectorAll("[id]"), function(newNode) {
        var id = getRawAttribute(newNode, "id");
        if (id && id.length > 0) {
          var normalizedId = id.replace("'", "\\'");
          var normalizedTag = newNode.tagName.replace(":", "\\:");
          var parentElt2 = asParentNode(parentNode);
          var oldNode = parentElt2 && parentElt2.querySelector(normalizedTag + "[id='" + normalizedId + "']");
          if (oldNode && oldNode !== parentElt2) {
            var newAttributes = newNode.cloneNode();
            cloneAttributes(newNode, oldNode);
            settleInfo.tasks.push(function() {
              cloneAttributes(newNode, newAttributes);
            });
          }
        }
      });
    }
    function makeAjaxLoadTask(child) {
      return function() {
        removeClassFromElement(child, htmx.config.addedClass);
        processNode(asElement(child));
        processFocus(asParentNode(child));
        triggerEvent(child, "htmx:load");
      };
    }
    function processFocus(child) {
      var autofocus = "[autofocus]";
      var autoFocusedElt = asHtmlElement(matches(child, autofocus) ? child : child.querySelector(autofocus));
      if (autoFocusedElt != null) {
        autoFocusedElt.focus();
      }
    }
    function insertNodesBefore(parentNode, insertBefore, fragment, settleInfo) {
      handleAttributes(parentNode, fragment, settleInfo);
      while (fragment.childNodes.length > 0) {
        var child = fragment.firstChild;
        addClassToElement(asElement(child), htmx.config.addedClass);
        parentNode.insertBefore(child, insertBefore);
        if (child.nodeType !== Node.TEXT_NODE && child.nodeType !== Node.COMMENT_NODE) {
          settleInfo.tasks.push(makeAjaxLoadTask(child));
        }
      }
    }
    function stringHash(string, hash) {
      var _char = 0;
      while (_char < string.length) {
        hash = (hash << 5) - hash + string.charCodeAt(_char++) | 0;
      }
      return hash;
    }
    function attributeHash(elt) {
      var hash = 0;
      if (elt.attributes) {
        for (var i = 0; i < elt.attributes.length; i++) {
          var attribute = elt.attributes[i];
          if (attribute.value) {
            hash = stringHash(attribute.name, hash);
            hash = stringHash(attribute.value, hash);
          }
        }
      }
      return hash;
    }
    function deInitOnHandlers(elt) {
      var internalData = getInternalData(elt);
      if (internalData.onHandlers) {
        for (var i = 0; i < internalData.onHandlers.length; i++) {
          var handlerInfo = internalData.onHandlers[i];
          removeEventListenerImpl(elt, handlerInfo.event, handlerInfo.listener);
        }
        delete internalData.onHandlers;
      }
    }
    function deInitNode(element) {
      var internalData = getInternalData(element);
      if (internalData.timeout) {
        clearTimeout(internalData.timeout);
      }
      if (internalData.listenerInfos) {
        forEach(internalData.listenerInfos, function(info) {
          if (info.on) {
            removeEventListenerImpl(info.on, info.trigger, info.listener);
          }
        });
      }
      deInitOnHandlers(element);
      forEach(Object.keys(internalData), function(key) {
        delete internalData[key];
      });
    }
    function cleanUpElement(element) {
      triggerEvent(element, "htmx:beforeCleanupElement");
      deInitNode(element);
      if (element.children) {
        forEach(element.children, function(child) {
          cleanUpElement(child);
        });
      }
    }
    function swapOuterHTML(target, fragment, settleInfo) {
      if (target instanceof Element && target.tagName === "BODY") {
        return swapInnerHTML(target, fragment, settleInfo);
      }
      var newElt;
      var eltBeforeNewContent = target.previousSibling;
      var parentNode = parentElt(target);
      if (!parentNode) {
        return;
      }
      insertNodesBefore(parentNode, target, fragment, settleInfo);
      if (eltBeforeNewContent == null) {
        newElt = parentNode.firstChild;
      } else {
        newElt = eltBeforeNewContent.nextSibling;
      }
      settleInfo.elts = settleInfo.elts.filter(function(e) {
        return e !== target;
      });
      while (newElt && newElt !== target) {
        if (newElt instanceof Element) {
          settleInfo.elts.push(newElt);
        }
        newElt = newElt.nextSibling;
      }
      cleanUpElement(target);
      if (target instanceof Element) {
        target.remove();
      } else {
        target.parentNode.removeChild(target);
      }
    }
    function swapAfterBegin(target, fragment, settleInfo) {
      return insertNodesBefore(target, target.firstChild, fragment, settleInfo);
    }
    function swapBeforeBegin(target, fragment, settleInfo) {
      return insertNodesBefore(parentElt(target), target, fragment, settleInfo);
    }
    function swapBeforeEnd(target, fragment, settleInfo) {
      return insertNodesBefore(target, null, fragment, settleInfo);
    }
    function swapAfterEnd(target, fragment, settleInfo) {
      return insertNodesBefore(parentElt(target), target.nextSibling, fragment, settleInfo);
    }
    function swapDelete(target) {
      cleanUpElement(target);
      var parent = parentElt(target);
      if (parent) {
        return parent.removeChild(target);
      }
    }
    function swapInnerHTML(target, fragment, settleInfo) {
      var firstChild = target.firstChild;
      insertNodesBefore(target, firstChild, fragment, settleInfo);
      if (firstChild) {
        while (firstChild.nextSibling) {
          cleanUpElement(firstChild.nextSibling);
          target.removeChild(firstChild.nextSibling);
        }
        cleanUpElement(firstChild);
        target.removeChild(firstChild);
      }
    }
    function swapWithStyle(swapStyle, elt, target, fragment, settleInfo) {
      switch (swapStyle) {
        case "none":
          return;
        case "outerHTML":
          swapOuterHTML(target, fragment, settleInfo);
          return;
        case "afterbegin":
          swapAfterBegin(target, fragment, settleInfo);
          return;
        case "beforebegin":
          swapBeforeBegin(target, fragment, settleInfo);
          return;
        case "beforeend":
          swapBeforeEnd(target, fragment, settleInfo);
          return;
        case "afterend":
          swapAfterEnd(target, fragment, settleInfo);
          return;
        case "delete":
          swapDelete(target);
          return;
        default:
          var extensions2 = getExtensions(elt);
          for (var i = 0; i < extensions2.length; i++) {
            var ext = extensions2[i];
            try {
              var newElements = ext.handleSwap(swapStyle, target, fragment, settleInfo);
              if (newElements) {
                if (Array.isArray(newElements)) {
                  for (var j = 0; j < newElements.length; j++) {
                    var child = newElements[j];
                    if (child.nodeType !== Node.TEXT_NODE && child.nodeType !== Node.COMMENT_NODE) {
                      settleInfo.tasks.push(makeAjaxLoadTask(child));
                    }
                  }
                }
                return;
              }
            } catch (e) {
              logError(e);
            }
          }
          if (swapStyle === "innerHTML") {
            swapInnerHTML(target, fragment, settleInfo);
          } else {
            swapWithStyle(htmx.config.defaultSwapStyle, elt, target, fragment, settleInfo);
          }
      }
    }
    function findAndSwapOobElements(fragment, settleInfo, rootNode) {
      var oobElts = findAll(fragment, "[hx-swap-oob], [data-hx-swap-oob]");
      forEach(oobElts, function(oobElement) {
        if (htmx.config.allowNestedOobSwaps || oobElement.parentElement === null) {
          var oobValue = getAttributeValue(oobElement, "hx-swap-oob");
          if (oobValue != null) {
            oobSwap(oobValue, oobElement, settleInfo, rootNode);
          }
        } else {
          oobElement.removeAttribute("hx-swap-oob");
          oobElement.removeAttribute("data-hx-swap-oob");
        }
      });
      return oobElts.length > 0;
    }
    function swap(target, content, swapSpec, swapOptions) {
      if (!swapOptions) {
        swapOptions = {};
      }
      target = resolveTarget(target);
      var rootNode = swapOptions.contextElement ? getRootNode(swapOptions.contextElement, false) : getDocument();
      var activeElt = document.activeElement;
      var selectionInfo = {};
      try {
        selectionInfo = {
          elt: activeElt,
          // @ts-ignore
          start: activeElt ? activeElt.selectionStart : null,
          // @ts-ignore
          end: activeElt ? activeElt.selectionEnd : null
        };
      } catch (e) {
      }
      var settleInfo = makeSettleInfo(target);
      if (swapSpec.swapStyle === "textContent") {
        target.textContent = content;
      } else {
        var fragment = makeFragment(content);
        settleInfo.title = fragment.title;
        if (swapOptions.selectOOB) {
          var oobSelectValues = swapOptions.selectOOB.split(",");
          for (var i = 0; i < oobSelectValues.length; i++) {
            var oobSelectValue = oobSelectValues[i].split(":", 2);
            var id = oobSelectValue[0].trim();
            if (id.indexOf("#") === 0) {
              id = id.substring(1);
            }
            var oobValue = oobSelectValue[1] || "true";
            var oobElement = fragment.querySelector("#" + id);
            if (oobElement) {
              oobSwap(oobValue, oobElement, settleInfo, rootNode);
            }
          }
        }
        findAndSwapOobElements(fragment, settleInfo, rootNode);
        forEach(
          findAll(fragment, "template"),
          /** @param {HTMLTemplateElement} template */
          function(template) {
            if (findAndSwapOobElements(template.content, settleInfo, rootNode)) {
              template.remove();
            }
          }
        );
        if (swapOptions.select) {
          var newFragment = getDocument().createDocumentFragment();
          forEach(fragment.querySelectorAll(swapOptions.select), function(node) {
            newFragment.appendChild(node);
          });
          fragment = newFragment;
        }
        handlePreservedElements(fragment);
        swapWithStyle(swapSpec.swapStyle, swapOptions.contextElement, target, fragment, settleInfo);
        restorePreservedElements();
      }
      if (selectionInfo.elt && !bodyContains(selectionInfo.elt) && getRawAttribute(selectionInfo.elt, "id")) {
        var newActiveElt = document.getElementById(getRawAttribute(selectionInfo.elt, "id"));
        var focusOptions = {
          preventScroll: swapSpec.focusScroll !== void 0 ? !swapSpec.focusScroll : !htmx.config.defaultFocusScroll
        };
        if (newActiveElt) {
          if (selectionInfo.start && newActiveElt.setSelectionRange) {
            try {
              newActiveElt.setSelectionRange(selectionInfo.start, selectionInfo.end);
            } catch (e) {
            }
          }
          newActiveElt.focus(focusOptions);
        }
      }
      target.classList.remove(htmx.config.swappingClass);
      forEach(settleInfo.elts, function(elt) {
        if (elt.classList) {
          elt.classList.add(htmx.config.settlingClass);
        }
        triggerEvent(elt, "htmx:afterSwap", swapOptions.eventInfo);
      });
      if (swapOptions.afterSwapCallback) {
        swapOptions.afterSwapCallback();
      }
      if (!swapSpec.ignoreTitle) {
        handleTitle(settleInfo.title);
      }
      var doSettle = function doSettle2() {
        forEach(settleInfo.tasks, function(task) {
          task.call();
        });
        forEach(settleInfo.elts, function(elt) {
          if (elt.classList) {
            elt.classList.remove(htmx.config.settlingClass);
          }
          triggerEvent(elt, "htmx:afterSettle", swapOptions.eventInfo);
        });
        if (swapOptions.anchor) {
          var anchorTarget = asElement(resolveTarget("#" + swapOptions.anchor));
          if (anchorTarget) {
            anchorTarget.scrollIntoView({
              block: "start",
              behavior: "auto"
            });
          }
        }
        updateScrollState(settleInfo.elts, swapSpec);
        if (swapOptions.afterSettleCallback) {
          swapOptions.afterSettleCallback();
        }
      };
      if (swapSpec.settleDelay > 0) {
        getWindow().setTimeout(doSettle, swapSpec.settleDelay);
      } else {
        doSettle();
      }
    }
    function handleTriggerHeader(xhr, header, elt) {
      var triggerBody = xhr.getResponseHeader(header);
      if (triggerBody.indexOf("{") === 0) {
        var triggers = parseJSON(triggerBody);
        for (var eventName in triggers) {
          if (triggers.hasOwnProperty(eventName)) {
            var detail = triggers[eventName];
            if (isRawObject(detail)) {
              elt = detail.target !== void 0 ? detail.target : elt;
            } else {
              detail = {
                value: detail
              };
            }
            triggerEvent(elt, eventName, detail);
          }
        }
      } else {
        var eventNames = triggerBody.split(",");
        for (var i = 0; i < eventNames.length; i++) {
          triggerEvent(elt, eventNames[i].trim(), []);
        }
      }
    }
    var WHITESPACE = /\s/;
    var WHITESPACE_OR_COMMA = /[\s,]/;
    var SYMBOL_START = /[_$a-zA-Z]/;
    var SYMBOL_CONT = /[_$a-zA-Z0-9]/;
    var STRINGISH_START = ['"', "'", "/"];
    var NOT_WHITESPACE = /[^\s]/;
    var COMBINED_SELECTOR_START = /[{(]/;
    var COMBINED_SELECTOR_END = /[})]/;
    function tokenizeString(str2) {
      var tokens = [];
      var position = 0;
      while (position < str2.length) {
        if (SYMBOL_START.exec(str2.charAt(position))) {
          var startPosition = position;
          while (SYMBOL_CONT.exec(str2.charAt(position + 1))) {
            position++;
          }
          tokens.push(str2.substr(startPosition, position - startPosition + 1));
        } else if (STRINGISH_START.indexOf(str2.charAt(position)) !== -1) {
          var startChar = str2.charAt(position);
          var startPosition = position;
          position++;
          while (position < str2.length && str2.charAt(position) !== startChar) {
            if (str2.charAt(position) === "\\") {
              position++;
            }
            position++;
          }
          tokens.push(str2.substr(startPosition, position - startPosition + 1));
        } else {
          var symbol = str2.charAt(position);
          tokens.push(symbol);
        }
        position++;
      }
      return tokens;
    }
    function isPossibleRelativeReference(token, last, paramName) {
      return SYMBOL_START.exec(token.charAt(0)) && token !== "true" && token !== "false" && token !== "this" && token !== paramName && last !== ".";
    }
    function maybeGenerateConditional(elt, tokens, paramName) {
      if (tokens[0] === "[") {
        tokens.shift();
        var bracketCount = 1;
        var conditionalSource = " return (function(" + paramName + "){ return (";
        var last = null;
        while (tokens.length > 0) {
          var token = tokens[0];
          if (token === "]") {
            bracketCount--;
            if (bracketCount === 0) {
              if (last === null) {
                conditionalSource = conditionalSource + "true";
              }
              tokens.shift();
              conditionalSource += ")})";
              try {
                var conditionFunction = maybeEval(elt, function() {
                  return Function(conditionalSource)();
                }, function() {
                  return true;
                });
                conditionFunction.source = conditionalSource;
                return conditionFunction;
              } catch (e) {
                triggerErrorEvent(getDocument().body, "htmx:syntax:error", {
                  error: e,
                  source: conditionalSource
                });
                return null;
              }
            }
          } else if (token === "[") {
            bracketCount++;
          }
          if (isPossibleRelativeReference(token, last, paramName)) {
            conditionalSource += "((" + paramName + "." + token + ") ? (" + paramName + "." + token + ") : (window." + token + "))";
          } else {
            conditionalSource = conditionalSource + token;
          }
          last = tokens.shift();
        }
      }
    }
    function consumeUntil(tokens, match) {
      var result = "";
      while (tokens.length > 0 && !match.test(tokens[0])) {
        result += tokens.shift();
      }
      return result;
    }
    function consumeCSSSelector(tokens) {
      var result;
      if (tokens.length > 0 && COMBINED_SELECTOR_START.test(tokens[0])) {
        tokens.shift();
        result = consumeUntil(tokens, COMBINED_SELECTOR_END).trim();
        tokens.shift();
      } else {
        result = consumeUntil(tokens, WHITESPACE_OR_COMMA);
      }
      return result;
    }
    var INPUT_SELECTOR = "input, textarea, select";
    function parseAndCacheTrigger(elt, explicitTrigger, cache) {
      var triggerSpecs = [];
      var tokens = tokenizeString(explicitTrigger);
      do {
        consumeUntil(tokens, NOT_WHITESPACE);
        var initialLength = tokens.length;
        var trigger = consumeUntil(tokens, /[,\[\s]/);
        if (trigger !== "") {
          if (trigger === "every") {
            var every = {
              trigger: "every"
            };
            consumeUntil(tokens, NOT_WHITESPACE);
            every.pollInterval = parseInterval(consumeUntil(tokens, /[,\[\s]/));
            consumeUntil(tokens, NOT_WHITESPACE);
            var eventFilter = maybeGenerateConditional(elt, tokens, "event");
            if (eventFilter) {
              every.eventFilter = eventFilter;
            }
            triggerSpecs.push(every);
          } else {
            var triggerSpec = {
              trigger: trigger
            };
            var eventFilter = maybeGenerateConditional(elt, tokens, "event");
            if (eventFilter) {
              triggerSpec.eventFilter = eventFilter;
            }
            consumeUntil(tokens, NOT_WHITESPACE);
            while (tokens.length > 0 && tokens[0] !== ",") {
              var token = tokens.shift();
              if (token === "changed") {
                triggerSpec.changed = true;
              } else if (token === "once") {
                triggerSpec.once = true;
              } else if (token === "consume") {
                triggerSpec.consume = true;
              } else if (token === "delay" && tokens[0] === ":") {
                tokens.shift();
                triggerSpec.delay = parseInterval(consumeUntil(tokens, WHITESPACE_OR_COMMA));
              } else if (token === "from" && tokens[0] === ":") {
                tokens.shift();
                if (COMBINED_SELECTOR_START.test(tokens[0])) {
                  var from_arg = consumeCSSSelector(tokens);
                } else {
                  var from_arg = consumeUntil(tokens, WHITESPACE_OR_COMMA);
                  if (from_arg === "closest" || from_arg === "find" || from_arg === "next" || from_arg === "previous") {
                    tokens.shift();
                    var selector = consumeCSSSelector(tokens);
                    if (selector.length > 0) {
                      from_arg += " " + selector;
                    }
                  }
                }
                triggerSpec.from = from_arg;
              } else if (token === "target" && tokens[0] === ":") {
                tokens.shift();
                triggerSpec.target = consumeCSSSelector(tokens);
              } else if (token === "throttle" && tokens[0] === ":") {
                tokens.shift();
                triggerSpec.throttle = parseInterval(consumeUntil(tokens, WHITESPACE_OR_COMMA));
              } else if (token === "queue" && tokens[0] === ":") {
                tokens.shift();
                triggerSpec.queue = consumeUntil(tokens, WHITESPACE_OR_COMMA);
              } else if (token === "root" && tokens[0] === ":") {
                tokens.shift();
                triggerSpec[token] = consumeCSSSelector(tokens);
              } else if (token === "threshold" && tokens[0] === ":") {
                tokens.shift();
                triggerSpec[token] = consumeUntil(tokens, WHITESPACE_OR_COMMA);
              } else {
                triggerErrorEvent(elt, "htmx:syntax:error", {
                  token: tokens.shift()
                });
              }
              consumeUntil(tokens, NOT_WHITESPACE);
            }
            triggerSpecs.push(triggerSpec);
          }
        }
        if (tokens.length === initialLength) {
          triggerErrorEvent(elt, "htmx:syntax:error", {
            token: tokens.shift()
          });
        }
        consumeUntil(tokens, NOT_WHITESPACE);
      } while (tokens[0] === "," && tokens.shift());
      if (cache) {
        cache[explicitTrigger] = triggerSpecs;
      }
      return triggerSpecs;
    }
    function getTriggerSpecs(elt) {
      var explicitTrigger = getAttributeValue(elt, "hx-trigger");
      var triggerSpecs = [];
      if (explicitTrigger) {
        var cache = htmx.config.triggerSpecsCache;
        triggerSpecs = cache && cache[explicitTrigger] || parseAndCacheTrigger(elt, explicitTrigger, cache);
      }
      if (triggerSpecs.length > 0) {
        return triggerSpecs;
      } else if (matches(elt, "form")) {
        return [{
          trigger: "submit"
        }];
      } else if (matches(elt, 'input[type="button"], input[type="submit"]')) {
        return [{
          trigger: "click"
        }];
      } else if (matches(elt, INPUT_SELECTOR)) {
        return [{
          trigger: "change"
        }];
      } else {
        return [{
          trigger: "click"
        }];
      }
    }
    function cancelPolling(elt) {
      getInternalData(elt).cancelled = true;
    }
    function processPolling(elt, handler, spec) {
      var nodeData = getInternalData(elt);
      nodeData.timeout = getWindow().setTimeout(function() {
        if (bodyContains(elt) && nodeData.cancelled !== true) {
          if (!maybeFilterEvent(spec, elt, makeEvent("hx:poll:trigger", {
            triggerSpec: spec,
            target: elt
          }))) {
            handler(elt);
          }
          processPolling(elt, handler, spec);
        }
      }, spec.pollInterval);
    }
    function isLocalLink(elt) {
      return location.hostname === elt.hostname && getRawAttribute(elt, "href") && getRawAttribute(elt, "href").indexOf("#") !== 0;
    }
    function eltIsDisabled(elt) {
      return closest(elt, htmx.config.disableSelector);
    }
    function boostElement(elt, nodeData, triggerSpecs) {
      if (elt instanceof HTMLAnchorElement && isLocalLink(elt) && (elt.target === "" || elt.target === "_self") || elt.tagName === "FORM" && String(getRawAttribute(elt, "method")).toLowerCase() !== "dialog") {
        nodeData.boosted = true;
        var verb, path;
        if (elt.tagName === "A") {
          verb = /** @type HttpVerb */
          "get";
          path = getRawAttribute(elt, "href");
        } else {
          var rawAttribute = getRawAttribute(elt, "method");
          verb = /** @type HttpVerb */
          rawAttribute ? rawAttribute.toLowerCase() : "get";
          path = getRawAttribute(elt, "action");
          if (verb === "get" && path.includes("?")) {
            path = path.replace(/\?[^#]+/, "");
          }
        }
        triggerSpecs.forEach(function(triggerSpec) {
          addEventListener(elt, function(node, evt) {
            var elt2 = asElement(node);
            if (eltIsDisabled(elt2)) {
              cleanUpElement(elt2);
              return;
            }
            issueAjaxRequest(verb, path, elt2, evt);
          }, nodeData, triggerSpec, true);
        });
      }
    }
    function shouldCancel(evt, node) {
      var elt = asElement(node);
      if (!elt) {
        return false;
      }
      if (evt.type === "submit" || evt.type === "click") {
        if (elt.tagName === "FORM") {
          return true;
        }
        if (matches(elt, 'input[type="submit"], button') && closest(elt, "form") !== null) {
          return true;
        }
        if (elt instanceof HTMLAnchorElement && elt.href && (elt.getAttribute("href") === "#" || elt.getAttribute("href").indexOf("#") !== 0)) {
          return true;
        }
      }
      return false;
    }
    function ignoreBoostedAnchorCtrlClick(elt, evt) {
      return getInternalData(elt).boosted && elt instanceof HTMLAnchorElement && evt.type === "click" && // @ts-ignore this will resolve to undefined for events that don't define those properties, which is fine
      (evt.ctrlKey || evt.metaKey);
    }
    function maybeFilterEvent(triggerSpec, elt, evt) {
      var eventFilter = triggerSpec.eventFilter;
      if (eventFilter) {
        try {
          return eventFilter.call(elt, evt) !== true;
        } catch (e) {
          var source = eventFilter.source;
          triggerErrorEvent(getDocument().body, "htmx:eventFilter:error", {
            error: e,
            source: source
          });
          return true;
        }
      }
      return false;
    }
    function addEventListener(elt, handler, nodeData, triggerSpec, explicitCancel) {
      var elementData = getInternalData(elt);
      var eltsToListenOn;
      if (triggerSpec.from) {
        eltsToListenOn = querySelectorAllExt(elt, triggerSpec.from);
      } else {
        eltsToListenOn = [elt];
      }
      if (triggerSpec.changed) {
        if (!("lastValue" in elementData)) {
          elementData.lastValue = /* @__PURE__ */ new WeakMap();
        }
        eltsToListenOn.forEach(function(eltToListenOn) {
          if (!elementData.lastValue.has(triggerSpec)) {
            elementData.lastValue.set(triggerSpec, /* @__PURE__ */ new WeakMap());
          }
          elementData.lastValue.get(triggerSpec).set(eltToListenOn, eltToListenOn.value);
        });
      }
      forEach(eltsToListenOn, function(eltToListenOn) {
        var _eventListener = function eventListener(evt) {
          if (!bodyContains(elt)) {
            eltToListenOn.removeEventListener(triggerSpec.trigger, _eventListener);
            return;
          }
          if (ignoreBoostedAnchorCtrlClick(elt, evt)) {
            return;
          }
          if (explicitCancel || shouldCancel(evt, elt)) {
            evt.preventDefault();
          }
          if (maybeFilterEvent(triggerSpec, elt, evt)) {
            return;
          }
          var eventData = getInternalData(evt);
          eventData.triggerSpec = triggerSpec;
          if (eventData.handledFor == null) {
            eventData.handledFor = [];
          }
          if (eventData.handledFor.indexOf(elt) < 0) {
            eventData.handledFor.push(elt);
            if (triggerSpec.consume) {
              evt.stopPropagation();
            }
            if (triggerSpec.target && evt.target) {
              if (!matches(asElement(evt.target), triggerSpec.target)) {
                return;
              }
            }
            if (triggerSpec.once) {
              if (elementData.triggeredOnce) {
                return;
              } else {
                elementData.triggeredOnce = true;
              }
            }
            if (triggerSpec.changed) {
              var node = event.target;
              var value = node.value;
              var lastValue = elementData.lastValue.get(triggerSpec);
              if (lastValue.has(node) && lastValue.get(node) === value) {
                return;
              }
              lastValue.set(node, value);
            }
            if (elementData.delayed) {
              clearTimeout(elementData.delayed);
            }
            if (elementData.throttle) {
              return;
            }
            if (triggerSpec.throttle > 0) {
              if (!elementData.throttle) {
                triggerEvent(elt, "htmx:trigger");
                handler(elt, evt);
                elementData.throttle = getWindow().setTimeout(function() {
                  elementData.throttle = null;
                }, triggerSpec.throttle);
              }
            } else if (triggerSpec.delay > 0) {
              elementData.delayed = getWindow().setTimeout(function() {
                triggerEvent(elt, "htmx:trigger");
                handler(elt, evt);
              }, triggerSpec.delay);
            } else {
              triggerEvent(elt, "htmx:trigger");
              handler(elt, evt);
            }
          }
        };
        if (nodeData.listenerInfos == null) {
          nodeData.listenerInfos = [];
        }
        nodeData.listenerInfos.push({
          trigger: triggerSpec.trigger,
          listener: _eventListener,
          on: eltToListenOn
        });
        eltToListenOn.addEventListener(triggerSpec.trigger, _eventListener);
      });
    }
    var windowIsScrolling = false;
    var scrollHandler = null;
    function initScrollHandler() {
      if (!scrollHandler) {
        scrollHandler = function scrollHandler2() {
          windowIsScrolling = true;
        };
        window.addEventListener("scroll", scrollHandler);
        window.addEventListener("resize", scrollHandler);
        setInterval(function() {
          if (windowIsScrolling) {
            windowIsScrolling = false;
            forEach(getDocument().querySelectorAll("[hx-trigger*='revealed'],[data-hx-trigger*='revealed']"), function(elt) {
              maybeReveal(elt);
            });
          }
        }, 200);
      }
    }
    function maybeReveal(elt) {
      if (!hasAttribute(elt, "data-hx-revealed") && isScrolledIntoView(elt)) {
        elt.setAttribute("data-hx-revealed", "true");
        var nodeData = getInternalData(elt);
        if (nodeData.initHash) {
          triggerEvent(elt, "revealed");
        } else {
          elt.addEventListener("htmx:afterProcessNode", function() {
            triggerEvent(elt, "revealed");
          }, {
            once: true
          });
        }
      }
    }
    function loadImmediately(elt, handler, nodeData, delay) {
      var load = function load2() {
        if (!nodeData.loaded) {
          nodeData.loaded = true;
          handler(elt);
        }
      };
      if (delay > 0) {
        getWindow().setTimeout(load, delay);
      } else {
        load();
      }
    }
    function processVerbs(elt, nodeData, triggerSpecs) {
      var explicitAction = false;
      forEach(VERBS, function(verb) {
        if (hasAttribute(elt, "hx-" + verb)) {
          var path = getAttributeValue(elt, "hx-" + verb);
          explicitAction = true;
          nodeData.path = path;
          nodeData.verb = verb;
          triggerSpecs.forEach(function(triggerSpec) {
            addTriggerHandler(elt, triggerSpec, nodeData, function(node, evt) {
              var elt2 = asElement(node);
              if (closest(elt2, htmx.config.disableSelector)) {
                cleanUpElement(elt2);
                return;
              }
              issueAjaxRequest(verb, path, elt2, evt);
            });
          });
        }
      });
      return explicitAction;
    }
    function addTriggerHandler(elt, triggerSpec, nodeData, handler) {
      if (triggerSpec.trigger === "revealed") {
        initScrollHandler();
        addEventListener(elt, handler, nodeData, triggerSpec);
        maybeReveal(asElement(elt));
      } else if (triggerSpec.trigger === "intersect") {
        var observerOptions = {};
        if (triggerSpec.root) {
          observerOptions.root = querySelectorExt(elt, triggerSpec.root);
        }
        if (triggerSpec.threshold) {
          observerOptions.threshold = parseFloat(triggerSpec.threshold);
        }
        var observer = new IntersectionObserver(function(entries) {
          for (var i = 0; i < entries.length; i++) {
            var entry = entries[i];
            if (entry.isIntersecting) {
              triggerEvent(elt, "intersect");
              break;
            }
          }
        }, observerOptions);
        observer.observe(asElement(elt));
        addEventListener(asElement(elt), handler, nodeData, triggerSpec);
      } else if (triggerSpec.trigger === "load") {
        if (!maybeFilterEvent(triggerSpec, elt, makeEvent("load", {
          elt: elt
        }))) {
          loadImmediately(asElement(elt), handler, nodeData, triggerSpec.delay);
        }
      } else if (triggerSpec.pollInterval > 0) {
        nodeData.polling = true;
        processPolling(asElement(elt), handler, triggerSpec);
      } else {
        addEventListener(elt, handler, nodeData, triggerSpec);
      }
    }
    function shouldProcessHxOn(node) {
      var elt = asElement(node);
      if (!elt) {
        return false;
      }
      var attributes = elt.attributes;
      for (var j = 0; j < attributes.length; j++) {
        var attrName = attributes[j].name;
        if (startsWith(attrName, "hx-on:") || startsWith(attrName, "data-hx-on:") || startsWith(attrName, "hx-on-") || startsWith(attrName, "data-hx-on-")) {
          return true;
        }
      }
      return false;
    }
    var HX_ON_QUERY = new XPathEvaluator().createExpression('.//*[@*[ starts-with(name(), "hx-on:") or starts-with(name(), "data-hx-on:") or starts-with(name(), "hx-on-") or starts-with(name(), "data-hx-on-") ]]');
    function processHXOnRoot(elt, elements) {
      if (shouldProcessHxOn(elt)) {
        elements.push(asElement(elt));
      }
      var iter = HX_ON_QUERY.evaluate(elt);
      var node = null;
      while (node = iter.iterateNext()) elements.push(asElement(node));
    }
    function findHxOnWildcardElements(elt) {
      var elements = [];
      if (elt instanceof DocumentFragment) {
        var _iterator2 = _createForOfIteratorHelper(elt.childNodes), _step2;
        try {
          for (_iterator2.s(); !(_step2 = _iterator2.n()).done; ) {
            var child = _step2.value;
            processHXOnRoot(child, elements);
          }
        } catch (err) {
          _iterator2.e(err);
        } finally {
          _iterator2.f();
        }
      } else {
        processHXOnRoot(elt, elements);
      }
      return elements;
    }
    function findElementsToProcess(elt) {
      if (elt.querySelectorAll) {
        var boostedSelector = ", [hx-boost] a, [data-hx-boost] a, a[hx-boost], a[data-hx-boost]";
        var extensionSelectors = [];
        for (var e in extensions) {
          var extension = extensions[e];
          if (extension.getSelectors) {
            var selectors = extension.getSelectors();
            if (selectors) {
              extensionSelectors.push(selectors);
            }
          }
        }
        var results = elt.querySelectorAll(VERB_SELECTOR + boostedSelector + ", form, [type='submit'], [hx-ext], [data-hx-ext], [hx-trigger], [data-hx-trigger]" + extensionSelectors.flat().map(function(s) {
          return ", " + s;
        }).join(""));
        return results;
      } else {
        return [];
      }
    }
    function maybeSetLastButtonClicked(evt) {
      var elt = (
        /** @type {HTMLButtonElement|HTMLInputElement} */
        closest(asElement(evt.target), "button, input[type='submit']")
      );
      var internalData = getRelatedFormData(evt);
      if (internalData) {
        internalData.lastButtonClicked = elt;
      }
    }
    function maybeUnsetLastButtonClicked(evt) {
      var internalData = getRelatedFormData(evt);
      if (internalData) {
        internalData.lastButtonClicked = null;
      }
    }
    function getRelatedFormData(evt) {
      var elt = closest(asElement(evt.target), "button, input[type='submit']");
      if (!elt) {
        return;
      }
      var form = resolveTarget("#" + getRawAttribute(elt, "form"), elt.getRootNode()) || closest(elt, "form");
      if (!form) {
        return;
      }
      return getInternalData(form);
    }
    function initButtonTracking(elt) {
      elt.addEventListener("click", maybeSetLastButtonClicked);
      elt.addEventListener("focusin", maybeSetLastButtonClicked);
      elt.addEventListener("focusout", maybeUnsetLastButtonClicked);
    }
    function addHxOnEventHandler(elt, eventName, code) {
      var nodeData = getInternalData(elt);
      if (!Array.isArray(nodeData.onHandlers)) {
        nodeData.onHandlers = [];
      }
      var func;
      var listener = function listener2(e) {
        maybeEval(elt, function() {
          if (eltIsDisabled(elt)) {
            return;
          }
          if (!func) {
            func = new Function("event", code);
          }
          func.call(elt, e);
        });
      };
      elt.addEventListener(eventName, listener);
      nodeData.onHandlers.push({
        event: eventName,
        listener: listener
      });
    }
    function processHxOnWildcard(elt) {
      deInitOnHandlers(elt);
      for (var i = 0; i < elt.attributes.length; i++) {
        var name = elt.attributes[i].name;
        var value = elt.attributes[i].value;
        if (startsWith(name, "hx-on") || startsWith(name, "data-hx-on")) {
          var afterOnPosition = name.indexOf("-on") + 3;
          var nextChar = name.slice(afterOnPosition, afterOnPosition + 1);
          if (nextChar === "-" || nextChar === ":") {
            var eventName = name.slice(afterOnPosition + 1);
            if (startsWith(eventName, ":")) {
              eventName = "htmx" + eventName;
            } else if (startsWith(eventName, "-")) {
              eventName = "htmx:" + eventName.slice(1);
            } else if (startsWith(eventName, "htmx-")) {
              eventName = "htmx:" + eventName.slice(5);
            }
            addHxOnEventHandler(elt, eventName, value);
          }
        }
      }
    }
    function initNode(elt) {
      if (closest(elt, htmx.config.disableSelector)) {
        cleanUpElement(elt);
        return;
      }
      var nodeData = getInternalData(elt);
      if (nodeData.initHash !== attributeHash(elt)) {
        deInitNode(elt);
        nodeData.initHash = attributeHash(elt);
        triggerEvent(elt, "htmx:beforeProcessNode");
        var triggerSpecs = getTriggerSpecs(elt);
        var hasExplicitHttpAction = processVerbs(elt, nodeData, triggerSpecs);
        if (!hasExplicitHttpAction) {
          if (getClosestAttributeValue(elt, "hx-boost") === "true") {
            boostElement(elt, nodeData, triggerSpecs);
          } else if (hasAttribute(elt, "hx-trigger")) {
            triggerSpecs.forEach(function(triggerSpec) {
              addTriggerHandler(elt, triggerSpec, nodeData, function() {
              });
            });
          }
        }
        if (elt.tagName === "FORM" || getRawAttribute(elt, "type") === "submit" && hasAttribute(elt, "form")) {
          initButtonTracking(elt);
        }
        triggerEvent(elt, "htmx:afterProcessNode");
      }
    }
    function processNode(elt) {
      elt = resolveTarget(elt);
      if (closest(elt, htmx.config.disableSelector)) {
        cleanUpElement(elt);
        return;
      }
      initNode(elt);
      forEach(findElementsToProcess(elt), function(child) {
        initNode(child);
      });
      forEach(findHxOnWildcardElements(elt), processHxOnWildcard);
    }
    function kebabEventName(str2) {
      return str2.replace(/([a-z0-9])([A-Z])/g, "$1-$2").toLowerCase();
    }
    function makeEvent(eventName, detail) {
      var evt;
      if (window.CustomEvent && typeof window.CustomEvent === "function") {
        evt = new CustomEvent(eventName, {
          bubbles: true,
          cancelable: true,
          composed: true,
          detail: detail
        });
      } else {
        evt = getDocument().createEvent("CustomEvent");
        evt.initCustomEvent(eventName, true, true, detail);
      }
      return evt;
    }
    function triggerErrorEvent(elt, eventName, detail) {
      triggerEvent(elt, eventName, mergeObjects({
        error: eventName
      }, detail));
    }
    function ignoreEventForLogging(eventName) {
      return eventName === "htmx:afterProcessNode";
    }
    function withExtensions(elt, toDo) {
      forEach(getExtensions(elt), function(extension) {
        try {
          toDo(extension);
        } catch (e) {
          logError(e);
        }
      });
    }
    function logError(msg) {
      if (console.error) {
        console.error(msg);
      } else if (console.log) {
        console.log("ERROR: ", msg);
      }
    }
    function triggerEvent(elt, eventName, detail) {
      elt = resolveTarget(elt);
      if (detail == null) {
        detail = {};
      }
      detail.elt = elt;
      var event2 = makeEvent(eventName, detail);
      if (htmx.logger && !ignoreEventForLogging(eventName)) {
        htmx.logger(elt, eventName, detail);
      }
      if (detail.error) {
        logError(detail.error);
        triggerEvent(elt, "htmx:error", {
          errorInfo: detail
        });
      }
      var eventResult = elt.dispatchEvent(event2);
      var kebabName = kebabEventName(eventName);
      if (eventResult && kebabName !== eventName) {
        var kebabedEvent = makeEvent(kebabName, event2.detail);
        eventResult = eventResult && elt.dispatchEvent(kebabedEvent);
      }
      withExtensions(asElement(elt), function(extension) {
        eventResult = eventResult && extension.onEvent(eventName, event2) !== false && !event2.defaultPrevented;
      });
      return eventResult;
    }
    var currentPathForHistory = location.pathname + location.search;
    function getHistoryElement() {
      var historyElt = getDocument().querySelector("[hx-history-elt],[data-hx-history-elt]");
      return historyElt || getDocument().body;
    }
    function saveToHistoryCache(url, rootElt) {
      if (!canAccessLocalStorage()) {
        return;
      }
      var innerHTML = cleanInnerHtmlForHistory(rootElt);
      var title = getDocument().title;
      var scroll = window.scrollY;
      if (htmx.config.historyCacheSize <= 0) {
        localStorage.removeItem("htmx-history-cache");
        return;
      }
      url = normalizePath(url);
      var historyCache = parseJSON(localStorage.getItem("htmx-history-cache")) || [];
      for (var i = 0; i < historyCache.length; i++) {
        if (historyCache[i].url === url) {
          historyCache.splice(i, 1);
          break;
        }
      }
      var newHistoryItem = {
        url: url,
        content: innerHTML,
        title: title,
        scroll: scroll
      };
      triggerEvent(getDocument().body, "htmx:historyItemCreated", {
        item: newHistoryItem,
        cache: historyCache
      });
      historyCache.push(newHistoryItem);
      while (historyCache.length > htmx.config.historyCacheSize) {
        historyCache.shift();
      }
      while (historyCache.length > 0) {
        try {
          localStorage.setItem("htmx-history-cache", JSON.stringify(historyCache));
          break;
        } catch (e) {
          triggerErrorEvent(getDocument().body, "htmx:historyCacheError", {
            cause: e,
            cache: historyCache
          });
          historyCache.shift();
        }
      }
    }
    function getCachedHistory(url) {
      if (!canAccessLocalStorage()) {
        return null;
      }
      url = normalizePath(url);
      var historyCache = parseJSON(localStorage.getItem("htmx-history-cache")) || [];
      for (var i = 0; i < historyCache.length; i++) {
        if (historyCache[i].url === url) {
          return historyCache[i];
        }
      }
      return null;
    }
    function cleanInnerHtmlForHistory(elt) {
      var className = htmx.config.requestClass;
      var clone = (
        /** @type Element */
        elt.cloneNode(true)
      );
      forEach(findAll(clone, "." + className), function(child) {
        removeClassFromElement(child, className);
      });
      forEach(findAll(clone, "[data-disabled-by-htmx]"), function(child) {
        child.removeAttribute("disabled");
      });
      return clone.innerHTML;
    }
    function saveCurrentPageToHistory() {
      var elt = getHistoryElement();
      var path = currentPathForHistory || location.pathname + location.search;
      var disableHistoryCache;
      try {
        disableHistoryCache = getDocument().querySelector('[hx-history="false" i],[data-hx-history="false" i]');
      } catch (e) {
        disableHistoryCache = getDocument().querySelector('[hx-history="false"],[data-hx-history="false"]');
      }
      if (!disableHistoryCache) {
        triggerEvent(getDocument().body, "htmx:beforeHistorySave", {
          path: path,
          historyElt: elt
        });
        saveToHistoryCache(path, elt);
      }
      if (htmx.config.historyEnabled) history.replaceState({
        htmx: true
      }, getDocument().title, window.location.href);
    }
    function pushUrlIntoHistory(path) {
      if (htmx.config.getCacheBusterParam) {
        path = path.replace(/org\.htmx\.cache-buster=[^&]*&?/, "");
        if (endsWith(path, "&") || endsWith(path, "?")) {
          path = path.slice(0, -1);
        }
      }
      if (htmx.config.historyEnabled) {
        history.pushState({
          htmx: true
        }, "", path);
      }
      currentPathForHistory = path;
    }
    function replaceUrlInHistory(path) {
      if (htmx.config.historyEnabled) history.replaceState({
        htmx: true
      }, "", path);
      currentPathForHistory = path;
    }
    function settleImmediately(tasks) {
      forEach(tasks, function(task) {
        task.call(void 0);
      });
    }
    function loadHistoryFromServer(path) {
      var request = new XMLHttpRequest();
      var details = {
        path: path,
        xhr: request
      };
      triggerEvent(getDocument().body, "htmx:historyCacheMiss", details);
      request.open("GET", path, true);
      request.setRequestHeader("HX-Request", "true");
      request.setRequestHeader("HX-History-Restore-Request", "true");
      request.setRequestHeader("HX-Current-URL", getDocument().location.href);
      request.onload = function() {
        if (this.status >= 200 && this.status < 400) {
          triggerEvent(getDocument().body, "htmx:historyCacheMissLoad", details);
          var fragment = makeFragment(this.response);
          var content = fragment.querySelector("[hx-history-elt],[data-hx-history-elt]") || fragment;
          var historyElement = getHistoryElement();
          var settleInfo = makeSettleInfo(historyElement);
          handleTitle(fragment.title);
          handlePreservedElements(fragment);
          swapInnerHTML(historyElement, content, settleInfo);
          restorePreservedElements();
          settleImmediately(settleInfo.tasks);
          currentPathForHistory = path;
          triggerEvent(getDocument().body, "htmx:historyRestore", {
            path: path,
            cacheMiss: true,
            serverResponse: this.response
          });
        } else {
          triggerErrorEvent(getDocument().body, "htmx:historyCacheMissLoadError", details);
        }
      };
      request.send();
    }
    function restoreHistory(path) {
      saveCurrentPageToHistory();
      path = path || location.pathname + location.search;
      var cached = getCachedHistory(path);
      if (cached) {
        var fragment = makeFragment(cached.content);
        var historyElement = getHistoryElement();
        var settleInfo = makeSettleInfo(historyElement);
        handleTitle(cached.title);
        handlePreservedElements(fragment);
        swapInnerHTML(historyElement, fragment, settleInfo);
        restorePreservedElements();
        settleImmediately(settleInfo.tasks);
        getWindow().setTimeout(function() {
          window.scrollTo(0, cached.scroll);
        }, 0);
        currentPathForHistory = path;
        triggerEvent(getDocument().body, "htmx:historyRestore", {
          path: path,
          item: cached
        });
      } else {
        if (htmx.config.refreshOnHistoryMiss) {
          window.location.reload(true);
        } else {
          loadHistoryFromServer(path);
        }
      }
    }
    function addRequestIndicatorClasses(elt) {
      var indicators = (
        /** @type Element[] */
        findAttributeTargets(elt, "hx-indicator")
      );
      if (indicators == null) {
        indicators = [elt];
      }
      forEach(indicators, function(ic) {
        var internalData = getInternalData(ic);
        internalData.requestCount = (internalData.requestCount || 0) + 1;
        ic.classList.add.call(ic.classList, htmx.config.requestClass);
      });
      return indicators;
    }
    function disableElements(elt) {
      var disabledElts = (
        /** @type Element[] */
        findAttributeTargets(elt, "hx-disabled-elt")
      );
      if (disabledElts == null) {
        disabledElts = [];
      }
      forEach(disabledElts, function(disabledElement) {
        var internalData = getInternalData(disabledElement);
        internalData.requestCount = (internalData.requestCount || 0) + 1;
        disabledElement.setAttribute("disabled", "");
        disabledElement.setAttribute("data-disabled-by-htmx", "");
      });
      return disabledElts;
    }
    function removeRequestIndicators(indicators, disabled) {
      forEach(indicators.concat(disabled), function(ele) {
        var internalData = getInternalData(ele);
        internalData.requestCount = (internalData.requestCount || 1) - 1;
      });
      forEach(indicators, function(ic) {
        var internalData = getInternalData(ic);
        if (internalData.requestCount === 0) {
          ic.classList.remove.call(ic.classList, htmx.config.requestClass);
        }
      });
      forEach(disabled, function(disabledElement) {
        var internalData = getInternalData(disabledElement);
        if (internalData.requestCount === 0) {
          disabledElement.removeAttribute("disabled");
          disabledElement.removeAttribute("data-disabled-by-htmx");
        }
      });
    }
    function haveSeenNode(processed, elt) {
      for (var i = 0; i < processed.length; i++) {
        var node = processed[i];
        if (node.isSameNode(elt)) {
          return true;
        }
      }
      return false;
    }
    function shouldInclude(element) {
      var elt = (
        /** @type {HTMLInputElement} */
        element
      );
      if (elt.name === "" || elt.name == null || elt.disabled || closest(elt, "fieldset[disabled]")) {
        return false;
      }
      if (elt.type === "button" || elt.type === "submit" || elt.tagName === "image" || elt.tagName === "reset" || elt.tagName === "file") {
        return false;
      }
      if (elt.type === "checkbox" || elt.type === "radio") {
        return elt.checked;
      }
      return true;
    }
    function addValueToFormData(name, value, formData) {
      if (name != null && value != null) {
        if (Array.isArray(value)) {
          value.forEach(function(v) {
            formData.append(name, v);
          });
        } else {
          formData.append(name, value);
        }
      }
    }
    function removeValueFromFormData(name, value, formData) {
      if (name != null && value != null) {
        var values = formData.getAll(name);
        if (Array.isArray(value)) {
          values = values.filter(function(v) {
            return value.indexOf(v) < 0;
          });
        } else {
          values = values.filter(function(v) {
            return v !== value;
          });
        }
        formData["delete"](name);
        forEach(values, function(v) {
          return formData.append(name, v);
        });
      }
    }
    function processInputValue(processed, formData, errors, elt, validate) {
      if (elt == null || haveSeenNode(processed, elt)) {
        return;
      } else {
        processed.push(elt);
      }
      if (shouldInclude(elt)) {
        var name = getRawAttribute(elt, "name");
        var value = elt.value;
        if (elt instanceof HTMLSelectElement && elt.multiple) {
          value = toArray(elt.querySelectorAll("option:checked")).map(function(e) {
            return (
              /** @type HTMLOptionElement */
              e.value
            );
          });
        }
        if (elt instanceof HTMLInputElement && elt.files) {
          value = toArray(elt.files);
        }
        addValueToFormData(name, value, formData);
        if (validate) {
          validateElement(elt, errors);
        }
      }
      if (elt instanceof HTMLFormElement) {
        forEach(elt.elements, function(input) {
          if (processed.indexOf(input) >= 0) {
            removeValueFromFormData(input.name, input.value, formData);
          } else {
            processed.push(input);
          }
          if (validate) {
            validateElement(input, errors);
          }
        });
        new FormData(elt).forEach(function(value2, name2) {
          if (value2 instanceof File && value2.name === "") {
            return;
          }
          addValueToFormData(name2, value2, formData);
        });
      }
    }
    function validateElement(elt, errors) {
      var element = (
        /** @type {HTMLElement & ElementInternals} */
        elt
      );
      if (element.willValidate) {
        triggerEvent(element, "htmx:validation:validate");
        if (!element.checkValidity()) {
          errors.push({
            elt: element,
            message: element.validationMessage,
            validity: element.validity
          });
          triggerEvent(element, "htmx:validation:failed", {
            message: element.validationMessage,
            validity: element.validity
          });
        }
      }
    }
    function overrideFormData(receiver, donor) {
      var _iterator3 = _createForOfIteratorHelper(donor.keys()), _step3;
      try {
        for (_iterator3.s(); !(_step3 = _iterator3.n()).done; ) {
          var key = _step3.value;
          receiver["delete"](key);
        }
      } catch (err) {
        _iterator3.e(err);
      } finally {
        _iterator3.f();
      }
      donor.forEach(function(value, key2) {
        receiver.append(key2, value);
      });
      return receiver;
    }
    function getInputValues(elt, verb) {
      var processed = [];
      var formData = new FormData();
      var priorityFormData = new FormData();
      var errors = [];
      var internalData = getInternalData(elt);
      if (internalData.lastButtonClicked && !bodyContains(internalData.lastButtonClicked)) {
        internalData.lastButtonClicked = null;
      }
      var validate = elt instanceof HTMLFormElement && elt.noValidate !== true || getAttributeValue(elt, "hx-validate") === "true";
      if (internalData.lastButtonClicked) {
        validate = validate && internalData.lastButtonClicked.formNoValidate !== true;
      }
      if (verb !== "get") {
        processInputValue(processed, priorityFormData, errors, closest(elt, "form"), validate);
      }
      processInputValue(processed, formData, errors, elt, validate);
      if (internalData.lastButtonClicked || elt.tagName === "BUTTON" || elt.tagName === "INPUT" && getRawAttribute(elt, "type") === "submit") {
        var button = internalData.lastButtonClicked || /** @type HTMLInputElement|HTMLButtonElement */
        elt;
        var name = getRawAttribute(button, "name");
        addValueToFormData(name, button.value, priorityFormData);
      }
      var includes = findAttributeTargets(elt, "hx-include");
      forEach(includes, function(node) {
        processInputValue(processed, formData, errors, asElement(node), validate);
        if (!matches(node, "form")) {
          forEach(asParentNode(node).querySelectorAll(INPUT_SELECTOR), function(descendant) {
            processInputValue(processed, formData, errors, descendant, validate);
          });
        }
      });
      overrideFormData(formData, priorityFormData);
      return {
        errors: errors,
        formData: formData,
        values: formDataProxy(formData)
      };
    }
    function appendParam(returnStr, name, realValue) {
      if (returnStr !== "") {
        returnStr += "&";
      }
      if (String(realValue) === "[object Object]") {
        realValue = JSON.stringify(realValue);
      }
      var s = encodeURIComponent(realValue);
      returnStr += encodeURIComponent(name) + "=" + s;
      return returnStr;
    }
    function urlEncode(values) {
      values = formDataFromObject(values);
      var returnStr = "";
      values.forEach(function(value, key) {
        returnStr = appendParam(returnStr, key, value);
      });
      return returnStr;
    }
    function getHeaders(elt, target, prompt2) {
      var headers = {
        "HX-Request": "true",
        "HX-Trigger": getRawAttribute(elt, "id"),
        "HX-Trigger-Name": getRawAttribute(elt, "name"),
        "HX-Target": getAttributeValue(target, "id"),
        "HX-Current-URL": getDocument().location.href
      };
      getValuesForElement(elt, "hx-headers", false, headers);
      if (prompt2 !== void 0) {
        headers["HX-Prompt"] = prompt2;
      }
      if (getInternalData(elt).boosted) {
        headers["HX-Boosted"] = "true";
      }
      return headers;
    }
    function filterValues(inputValues, elt) {
      var paramsValue = getClosestAttributeValue(elt, "hx-params");
      if (paramsValue) {
        if (paramsValue === "none") {
          return new FormData();
        } else if (paramsValue === "*") {
          return inputValues;
        } else if (paramsValue.indexOf("not ") === 0) {
          forEach(paramsValue.substr(4).split(","), function(name) {
            name = name.trim();
            inputValues["delete"](name);
          });
          return inputValues;
        } else {
          var newValues = new FormData();
          forEach(paramsValue.split(","), function(name) {
            name = name.trim();
            if (inputValues.has(name)) {
              inputValues.getAll(name).forEach(function(value) {
                newValues.append(name, value);
              });
            }
          });
          return newValues;
        }
      } else {
        return inputValues;
      }
    }
    function isAnchorLink(elt) {
      return !!getRawAttribute(elt, "href") && getRawAttribute(elt, "href").indexOf("#") >= 0;
    }
    function getSwapSpecification(elt, swapInfoOverride) {
      var swapInfo = swapInfoOverride || getClosestAttributeValue(elt, "hx-swap");
      var swapSpec = {
        swapStyle: getInternalData(elt).boosted ? "innerHTML" : htmx.config.defaultSwapStyle,
        swapDelay: htmx.config.defaultSwapDelay,
        settleDelay: htmx.config.defaultSettleDelay
      };
      if (htmx.config.scrollIntoViewOnBoost && getInternalData(elt).boosted && !isAnchorLink(elt)) {
        swapSpec.show = "top";
      }
      if (swapInfo) {
        var split = splitOnWhitespace(swapInfo);
        if (split.length > 0) {
          for (var i = 0; i < split.length; i++) {
            var value = split[i];
            if (value.indexOf("swap:") === 0) {
              swapSpec.swapDelay = parseInterval(value.substr(5));
            } else if (value.indexOf("settle:") === 0) {
              swapSpec.settleDelay = parseInterval(value.substr(7));
            } else if (value.indexOf("transition:") === 0) {
              swapSpec.transition = value.substr(11) === "true";
            } else if (value.indexOf("ignoreTitle:") === 0) {
              swapSpec.ignoreTitle = value.substr(12) === "true";
            } else if (value.indexOf("scroll:") === 0) {
              var scrollSpec = value.substr(7);
              var splitSpec = scrollSpec.split(":");
              var scrollVal = splitSpec.pop();
              var selectorVal = splitSpec.length > 0 ? splitSpec.join(":") : null;
              swapSpec.scroll = scrollVal;
              swapSpec.scrollTarget = selectorVal;
            } else if (value.indexOf("show:") === 0) {
              var showSpec = value.substr(5);
              var splitSpec = showSpec.split(":");
              var showVal = splitSpec.pop();
              var selectorVal = splitSpec.length > 0 ? splitSpec.join(":") : null;
              swapSpec.show = showVal;
              swapSpec.showTarget = selectorVal;
            } else if (value.indexOf("focus-scroll:") === 0) {
              var focusScrollVal = value.substr("focus-scroll:".length);
              swapSpec.focusScroll = focusScrollVal == "true";
            } else if (i == 0) {
              swapSpec.swapStyle = value;
            } else {
              logError("Unknown modifier in hx-swap: " + value);
            }
          }
        }
      }
      return swapSpec;
    }
    function usesFormData(elt) {
      return getClosestAttributeValue(elt, "hx-encoding") === "multipart/form-data" || matches(elt, "form") && getRawAttribute(elt, "enctype") === "multipart/form-data";
    }
    function encodeParamsForBody(xhr, elt, filteredParameters) {
      var encodedParameters = null;
      withExtensions(elt, function(extension) {
        if (encodedParameters == null) {
          encodedParameters = extension.encodeParameters(xhr, filteredParameters, elt);
        }
      });
      if (encodedParameters != null) {
        return encodedParameters;
      } else {
        if (usesFormData(elt)) {
          return overrideFormData(new FormData(), formDataFromObject(filteredParameters));
        } else {
          return urlEncode(filteredParameters);
        }
      }
    }
    function makeSettleInfo(target) {
      return {
        tasks: [],
        elts: [target]
      };
    }
    function updateScrollState(content, swapSpec) {
      var first = content[0];
      var last = content[content.length - 1];
      if (swapSpec.scroll) {
        var target = null;
        if (swapSpec.scrollTarget) {
          target = asElement(querySelectorExt(first, swapSpec.scrollTarget));
        }
        if (swapSpec.scroll === "top" && (first || target)) {
          target = target || first;
          target.scrollTop = 0;
        }
        if (swapSpec.scroll === "bottom" && (last || target)) {
          target = target || last;
          target.scrollTop = target.scrollHeight;
        }
      }
      if (swapSpec.show) {
        var target = null;
        if (swapSpec.showTarget) {
          var targetStr = swapSpec.showTarget;
          if (swapSpec.showTarget === "window") {
            targetStr = "body";
          }
          target = asElement(querySelectorExt(first, targetStr));
        }
        if (swapSpec.show === "top" && (first || target)) {
          target = target || first;
          target.scrollIntoView({
            block: "start",
            behavior: htmx.config.scrollBehavior
          });
        }
        if (swapSpec.show === "bottom" && (last || target)) {
          target = target || last;
          target.scrollIntoView({
            block: "end",
            behavior: htmx.config.scrollBehavior
          });
        }
      }
    }
    function getValuesForElement(elt, attr, evalAsDefault, values) {
      if (values == null) {
        values = {};
      }
      if (elt == null) {
        return values;
      }
      var attributeValue = getAttributeValue(elt, attr);
      if (attributeValue) {
        var str2 = attributeValue.trim();
        var evaluateValue = evalAsDefault;
        if (str2 === "unset") {
          return null;
        }
        if (str2.indexOf("javascript:") === 0) {
          str2 = str2.substr(11);
          evaluateValue = true;
        } else if (str2.indexOf("js:") === 0) {
          str2 = str2.substr(3);
          evaluateValue = true;
        }
        if (str2.indexOf("{") !== 0) {
          str2 = "{" + str2 + "}";
        }
        var varsValues;
        if (evaluateValue) {
          varsValues = maybeEval(elt, function() {
            return Function("return (" + str2 + ")")();
          }, {});
        } else {
          varsValues = parseJSON(str2);
        }
        for (var key in varsValues) {
          if (varsValues.hasOwnProperty(key)) {
            if (values[key] == null) {
              values[key] = varsValues[key];
            }
          }
        }
      }
      return getValuesForElement(asElement(parentElt(elt)), attr, evalAsDefault, values);
    }
    function maybeEval(elt, toEval, defaultVal) {
      if (htmx.config.allowEval) {
        return toEval();
      } else {
        triggerErrorEvent(elt, "htmx:evalDisallowedError");
        return defaultVal;
      }
    }
    function getHXVarsForElement(elt, expressionVars) {
      return getValuesForElement(elt, "hx-vars", true, expressionVars);
    }
    function getHXValsForElement(elt, expressionVars) {
      return getValuesForElement(elt, "hx-vals", false, expressionVars);
    }
    function getExpressionVars(elt) {
      return mergeObjects(getHXVarsForElement(elt), getHXValsForElement(elt));
    }
    function safelySetHeaderValue(xhr, header, headerValue) {
      if (headerValue !== null) {
        try {
          xhr.setRequestHeader(header, headerValue);
        } catch (e) {
          xhr.setRequestHeader(header, encodeURIComponent(headerValue));
          xhr.setRequestHeader(header + "-URI-AutoEncoded", "true");
        }
      }
    }
    function getPathFromResponse(xhr) {
      if (xhr.responseURL && typeof URL !== "undefined") {
        try {
          var url = new URL(xhr.responseURL);
          return url.pathname + url.search;
        } catch (e) {
          triggerErrorEvent(getDocument().body, "htmx:badResponseUrl", {
            url: xhr.responseURL
          });
        }
      }
    }
    function hasHeader(xhr, regexp) {
      return regexp.test(xhr.getAllResponseHeaders());
    }
    function ajaxHelper(verb, path, context) {
      verb = /** @type HttpVerb */
      verb.toLowerCase();
      if (context) {
        if (context instanceof Element || typeof context === "string") {
          return issueAjaxRequest(verb, path, null, null, {
            targetOverride: resolveTarget(context) || DUMMY_ELT,
            returnPromise: true
          });
        } else {
          var resolvedTarget = resolveTarget(context.target);
          if (context.target && !resolvedTarget || !resolvedTarget && !resolveTarget(context.source)) {
            resolvedTarget = DUMMY_ELT;
          }
          return issueAjaxRequest(verb, path, resolveTarget(context.source), context.event, {
            handler: context.handler,
            headers: context.headers,
            values: context.values,
            targetOverride: resolvedTarget,
            swapOverride: context.swap,
            select: context.select,
            returnPromise: true
          });
        }
      } else {
        return issueAjaxRequest(verb, path, null, null, {
          returnPromise: true
        });
      }
    }
    function hierarchyForElt(elt) {
      var arr = [];
      while (elt) {
        arr.push(elt);
        elt = elt.parentElement;
      }
      return arr;
    }
    function verifyPath(elt, path, requestConfig) {
      var sameHost;
      var url;
      if (typeof URL === "function") {
        url = new URL(path, document.location.href);
        var origin = document.location.origin;
        sameHost = origin === url.origin;
      } else {
        url = path;
        sameHost = startsWith(path, document.location.origin);
      }
      if (htmx.config.selfRequestsOnly) {
        if (!sameHost) {
          return false;
        }
      }
      return triggerEvent(elt, "htmx:validateUrl", mergeObjects({
        url: url,
        sameHost: sameHost
      }, requestConfig));
    }
    function formDataFromObject(obj) {
      if (obj instanceof FormData) return obj;
      var formData = new FormData();
      var _loop2 = function _loop22(key2) {
        if (obj.hasOwnProperty(key2)) {
          if (obj[key2] && typeof obj[key2].forEach === "function") {
            obj[key2].forEach(function(v) {
              formData.append(key2, v);
            });
          } else if (_typeof(obj[key2]) === "object" && !(obj[key2] instanceof Blob)) {
            formData.append(key2, JSON.stringify(obj[key2]));
          } else {
            formData.append(key2, obj[key2]);
          }
        }
      };
      for (var key in obj) {
        _loop2(key);
      }
      return formData;
    }
    function formDataArrayProxy(formData, name, array) {
      return new Proxy(array, {
        get: function get(target, key) {
          if (typeof key === "number") return target[key];
          if (key === "length") return target.length;
          if (key === "push") {
            return function(value) {
              target.push(value);
              formData.append(name, value);
            };
          }
          if (typeof target[key] === "function") {
            return function() {
              target[key].apply(target, arguments);
              formData["delete"](name);
              target.forEach(function(v) {
                formData.append(name, v);
              });
            };
          }
          if (target[key] && target[key].length === 1) {
            return target[key][0];
          } else {
            return target[key];
          }
        },
        set: function set(target, index, value) {
          target[index] = value;
          formData["delete"](name);
          target.forEach(function(v) {
            formData.append(name, v);
          });
          return true;
        }
      });
    }
    function formDataProxy(formData) {
      return new Proxy(formData, {
        get: function get(target, name) {
          if (_typeof(name) === "symbol") {
            return Reflect.get(target, name);
          }
          if (name === "toJSON") {
            return function() {
              return Object.fromEntries(formData);
            };
          }
          if (name in target) {
            if (typeof target[name] === "function") {
              return function() {
                return formData[name].apply(formData, arguments);
              };
            } else {
              return target[name];
            }
          }
          var array = formData.getAll(name);
          if (array.length === 0) {
            return void 0;
          } else if (array.length === 1) {
            return array[0];
          } else {
            return formDataArrayProxy(target, name, array);
          }
        },
        set: function set(target, name, value) {
          if (typeof name !== "string") {
            return false;
          }
          target["delete"](name);
          if (value && typeof value.forEach === "function") {
            value.forEach(function(v) {
              target.append(name, v);
            });
          } else if (_typeof(value) === "object" && !(value instanceof Blob)) {
            target.append(name, JSON.stringify(value));
          } else {
            target.append(name, value);
          }
          return true;
        },
        deleteProperty: function deleteProperty(target, name) {
          if (typeof name === "string") {
            target["delete"](name);
          }
          return true;
        },
        // Support Object.assign call from proxy
        ownKeys: function ownKeys(target) {
          return Reflect.ownKeys(Object.fromEntries(target));
        },
        getOwnPropertyDescriptor: function getOwnPropertyDescriptor(target, prop) {
          return Reflect.getOwnPropertyDescriptor(Object.fromEntries(target), prop);
        }
      });
    }
    function issueAjaxRequest(verb, path, elt, event2, etc, confirmed) {
      var resolve = null;
      var reject = null;
      etc = etc != null ? etc : {};
      if (etc.returnPromise && typeof Promise !== "undefined") {
        var promise = new Promise(function(_resolve, _reject) {
          resolve = _resolve;
          reject = _reject;
        });
      }
      if (elt == null) {
        elt = getDocument().body;
      }
      var responseHandler = etc.handler || handleAjaxResponse;
      var select = etc.select || null;
      if (!bodyContains(elt)) {
        maybeCall(resolve);
        return promise;
      }
      var target = etc.targetOverride || asElement(getTarget(elt));
      if (target == null || target == DUMMY_ELT) {
        triggerErrorEvent(elt, "htmx:targetError", {
          target: getAttributeValue(elt, "hx-target")
        });
        maybeCall(reject);
        return promise;
      }
      var eltData = getInternalData(elt);
      var submitter = eltData.lastButtonClicked;
      if (submitter) {
        var buttonPath = getRawAttribute(submitter, "formaction");
        if (buttonPath != null) {
          path = buttonPath;
        }
        var buttonVerb = getRawAttribute(submitter, "formmethod");
        if (buttonVerb != null) {
          if (buttonVerb.toLowerCase() !== "dialog") {
            verb = /** @type HttpVerb */
            buttonVerb;
          }
        }
      }
      var confirmQuestion = getClosestAttributeValue(elt, "hx-confirm");
      if (confirmed === void 0) {
        var issueRequest = function issueRequest2(skipConfirmation) {
          return issueAjaxRequest(verb, path, elt, event2, etc, !!skipConfirmation);
        };
        var confirmDetails = {
          target: target,
          elt: elt,
          path: path,
          verb: verb,
          triggeringEvent: event2,
          etc: etc,
          issueRequest: issueRequest,
          question: confirmQuestion
        };
        if (triggerEvent(elt, "htmx:confirm", confirmDetails) === false) {
          maybeCall(resolve);
          return promise;
        }
      }
      var syncElt = elt;
      var syncStrategy = getClosestAttributeValue(elt, "hx-sync");
      var queueStrategy = null;
      var abortable = false;
      if (syncStrategy) {
        var syncStrings = syncStrategy.split(":");
        var selector = syncStrings[0].trim();
        if (selector === "this") {
          syncElt = findThisElement(elt, "hx-sync");
        } else {
          syncElt = asElement(querySelectorExt(elt, selector));
        }
        syncStrategy = (syncStrings[1] || "drop").trim();
        eltData = getInternalData(syncElt);
        if (syncStrategy === "drop" && eltData.xhr && eltData.abortable !== true) {
          maybeCall(resolve);
          return promise;
        } else if (syncStrategy === "abort") {
          if (eltData.xhr) {
            maybeCall(resolve);
            return promise;
          } else {
            abortable = true;
          }
        } else if (syncStrategy === "replace") {
          triggerEvent(syncElt, "htmx:abort");
        } else if (syncStrategy.indexOf("queue") === 0) {
          var queueStrArray = syncStrategy.split(" ");
          queueStrategy = (queueStrArray[1] || "last").trim();
        }
      }
      if (eltData.xhr) {
        if (eltData.abortable) {
          triggerEvent(syncElt, "htmx:abort");
        } else {
          if (queueStrategy == null) {
            if (event2) {
              var eventData = getInternalData(event2);
              if (eventData && eventData.triggerSpec && eventData.triggerSpec.queue) {
                queueStrategy = eventData.triggerSpec.queue;
              }
            }
            if (queueStrategy == null) {
              queueStrategy = "last";
            }
          }
          if (eltData.queuedRequests == null) {
            eltData.queuedRequests = [];
          }
          if (queueStrategy === "first" && eltData.queuedRequests.length === 0) {
            eltData.queuedRequests.push(function() {
              issueAjaxRequest(verb, path, elt, event2, etc);
            });
          } else if (queueStrategy === "all") {
            eltData.queuedRequests.push(function() {
              issueAjaxRequest(verb, path, elt, event2, etc);
            });
          } else if (queueStrategy === "last") {
            eltData.queuedRequests = [];
            eltData.queuedRequests.push(function() {
              issueAjaxRequest(verb, path, elt, event2, etc);
            });
          }
          maybeCall(resolve);
          return promise;
        }
      }
      var xhr = new XMLHttpRequest();
      eltData.xhr = xhr;
      eltData.abortable = abortable;
      var endRequestLock = function endRequestLock2() {
        eltData.xhr = null;
        eltData.abortable = false;
        if (eltData.queuedRequests != null && eltData.queuedRequests.length > 0) {
          var queuedRequest = eltData.queuedRequests.shift();
          queuedRequest();
        }
      };
      var promptQuestion = getClosestAttributeValue(elt, "hx-prompt");
      if (promptQuestion) {
        var promptResponse = prompt(promptQuestion);
        if (promptResponse === null || !triggerEvent(elt, "htmx:prompt", {
          prompt: promptResponse,
          target: target
        })) {
          maybeCall(resolve);
          endRequestLock();
          return promise;
        }
      }
      if (confirmQuestion && !confirmed) {
        if (!confirm(confirmQuestion)) {
          maybeCall(resolve);
          endRequestLock();
          return promise;
        }
      }
      var headers = getHeaders(elt, target, promptResponse);
      if (verb !== "get" && !usesFormData(elt)) {
        headers["Content-Type"] = "application/x-www-form-urlencoded";
      }
      if (etc.headers) {
        headers = mergeObjects(headers, etc.headers);
      }
      var results = getInputValues(elt, verb);
      var errors = results.errors;
      var rawFormData = results.formData;
      if (etc.values) {
        overrideFormData(rawFormData, formDataFromObject(etc.values));
      }
      var expressionVars = formDataFromObject(getExpressionVars(elt));
      var allFormData = overrideFormData(rawFormData, expressionVars);
      var filteredFormData = filterValues(allFormData, elt);
      if (htmx.config.getCacheBusterParam && verb === "get") {
        filteredFormData.set("org.htmx.cache-buster", getRawAttribute(target, "id") || "true");
      }
      if (path == null || path === "") {
        path = getDocument().location.href;
      }
      var requestAttrValues = getValuesForElement(elt, "hx-request");
      var eltIsBoosted = getInternalData(elt).boosted;
      var useUrlParams = htmx.config.methodsThatUseUrlParams.indexOf(verb) >= 0;
      var requestConfig = {
        boosted: eltIsBoosted,
        useUrlParams: useUrlParams,
        formData: filteredFormData,
        parameters: formDataProxy(filteredFormData),
        unfilteredFormData: allFormData,
        unfilteredParameters: formDataProxy(allFormData),
        headers: headers,
        target: target,
        verb: verb,
        errors: errors,
        withCredentials: etc.credentials || requestAttrValues.credentials || htmx.config.withCredentials,
        timeout: etc.timeout || requestAttrValues.timeout || htmx.config.timeout,
        path: path,
        triggeringEvent: event2
      };
      if (!triggerEvent(elt, "htmx:configRequest", requestConfig)) {
        maybeCall(resolve);
        endRequestLock();
        return promise;
      }
      path = requestConfig.path;
      verb = requestConfig.verb;
      headers = requestConfig.headers;
      filteredFormData = formDataFromObject(requestConfig.parameters);
      errors = requestConfig.errors;
      useUrlParams = requestConfig.useUrlParams;
      if (errors && errors.length > 0) {
        triggerEvent(elt, "htmx:validation:halted", requestConfig);
        maybeCall(resolve);
        endRequestLock();
        return promise;
      }
      var splitPath = path.split("#");
      var pathNoAnchor = splitPath[0];
      var anchor = splitPath[1];
      var finalPath = path;
      if (useUrlParams) {
        finalPath = pathNoAnchor;
        var hasValues = !filteredFormData.keys().next().done;
        if (hasValues) {
          if (finalPath.indexOf("?") < 0) {
            finalPath += "?";
          } else {
            finalPath += "&";
          }
          finalPath += urlEncode(filteredFormData);
          if (anchor) {
            finalPath += "#" + anchor;
          }
        }
      }
      if (!verifyPath(elt, finalPath, requestConfig)) {
        triggerErrorEvent(elt, "htmx:invalidPath", requestConfig);
        maybeCall(reject);
        return promise;
      }
      xhr.open(verb.toUpperCase(), finalPath, true);
      xhr.overrideMimeType("text/html");
      xhr.withCredentials = requestConfig.withCredentials;
      xhr.timeout = requestConfig.timeout;
      if (requestAttrValues.noHeaders) {
      } else {
        for (var header in headers) {
          if (headers.hasOwnProperty(header)) {
            var headerValue = headers[header];
            safelySetHeaderValue(xhr, header, headerValue);
          }
        }
      }
      var responseInfo = {
        xhr: xhr,
        target: target,
        requestConfig: requestConfig,
        etc: etc,
        boosted: eltIsBoosted,
        select: select,
        pathInfo: {
          requestPath: path,
          finalRequestPath: finalPath,
          responsePath: null,
          anchor: anchor
        }
      };
      xhr.onload = function() {
        try {
          var hierarchy = hierarchyForElt(elt);
          responseInfo.pathInfo.responsePath = getPathFromResponse(xhr);
          responseHandler(elt, responseInfo);
          if (responseInfo.keepIndicators !== true) {
            removeRequestIndicators(indicators, disableElts);
          }
          triggerEvent(elt, "htmx:afterRequest", responseInfo);
          triggerEvent(elt, "htmx:afterOnLoad", responseInfo);
          if (!bodyContains(elt)) {
            var secondaryTriggerElt = null;
            while (hierarchy.length > 0 && secondaryTriggerElt == null) {
              var parentEltInHierarchy = hierarchy.shift();
              if (bodyContains(parentEltInHierarchy)) {
                secondaryTriggerElt = parentEltInHierarchy;
              }
            }
            if (secondaryTriggerElt) {
              triggerEvent(secondaryTriggerElt, "htmx:afterRequest", responseInfo);
              triggerEvent(secondaryTriggerElt, "htmx:afterOnLoad", responseInfo);
            }
          }
          maybeCall(resolve);
          endRequestLock();
        } catch (e) {
          triggerErrorEvent(elt, "htmx:onLoadError", mergeObjects({
            error: e
          }, responseInfo));
          throw e;
        }
      };
      xhr.onerror = function() {
        removeRequestIndicators(indicators, disableElts);
        triggerErrorEvent(elt, "htmx:afterRequest", responseInfo);
        triggerErrorEvent(elt, "htmx:sendError", responseInfo);
        maybeCall(reject);
        endRequestLock();
      };
      xhr.onabort = function() {
        removeRequestIndicators(indicators, disableElts);
        triggerErrorEvent(elt, "htmx:afterRequest", responseInfo);
        triggerErrorEvent(elt, "htmx:sendAbort", responseInfo);
        maybeCall(reject);
        endRequestLock();
      };
      xhr.ontimeout = function() {
        removeRequestIndicators(indicators, disableElts);
        triggerErrorEvent(elt, "htmx:afterRequest", responseInfo);
        triggerErrorEvent(elt, "htmx:timeout", responseInfo);
        maybeCall(reject);
        endRequestLock();
      };
      if (!triggerEvent(elt, "htmx:beforeRequest", responseInfo)) {
        maybeCall(resolve);
        endRequestLock();
        return promise;
      }
      var indicators = addRequestIndicatorClasses(elt);
      var disableElts = disableElements(elt);
      forEach(["loadstart", "loadend", "progress", "abort"], function(eventName) {
        forEach([xhr, xhr.upload], function(target2) {
          target2.addEventListener(eventName, function(event3) {
            triggerEvent(elt, "htmx:xhr:" + eventName, {
              lengthComputable: event3.lengthComputable,
              loaded: event3.loaded,
              total: event3.total
            });
          });
        });
      });
      triggerEvent(elt, "htmx:beforeSend", responseInfo);
      var params = useUrlParams ? null : encodeParamsForBody(xhr, elt, filteredFormData);
      xhr.send(params);
      return promise;
    }
    function determineHistoryUpdates(elt, responseInfo) {
      var xhr = responseInfo.xhr;
      var pathFromHeaders = null;
      var typeFromHeaders = null;
      if (hasHeader(xhr, /HX-Push:/i)) {
        pathFromHeaders = xhr.getResponseHeader("HX-Push");
        typeFromHeaders = "push";
      } else if (hasHeader(xhr, /HX-Push-Url:/i)) {
        pathFromHeaders = xhr.getResponseHeader("HX-Push-Url");
        typeFromHeaders = "push";
      } else if (hasHeader(xhr, /HX-Replace-Url:/i)) {
        pathFromHeaders = xhr.getResponseHeader("HX-Replace-Url");
        typeFromHeaders = "replace";
      }
      if (pathFromHeaders) {
        if (pathFromHeaders === "false") {
          return {};
        } else {
          return {
            type: typeFromHeaders,
            path: pathFromHeaders
          };
        }
      }
      var requestPath = responseInfo.pathInfo.finalRequestPath;
      var responsePath = responseInfo.pathInfo.responsePath;
      var pushUrl = getClosestAttributeValue(elt, "hx-push-url");
      var replaceUrl = getClosestAttributeValue(elt, "hx-replace-url");
      var elementIsBoosted = getInternalData(elt).boosted;
      var saveType = null;
      var path = null;
      if (pushUrl) {
        saveType = "push";
        path = pushUrl;
      } else if (replaceUrl) {
        saveType = "replace";
        path = replaceUrl;
      } else if (elementIsBoosted) {
        saveType = "push";
        path = responsePath || requestPath;
      }
      if (path) {
        if (path === "false") {
          return {};
        }
        if (path === "true") {
          path = responsePath || requestPath;
        }
        if (responseInfo.pathInfo.anchor && path.indexOf("#") === -1) {
          path = path + "#" + responseInfo.pathInfo.anchor;
        }
        return {
          type: saveType,
          path: path
        };
      } else {
        return {};
      }
    }
    function codeMatches(responseHandlingConfig, status) {
      var regExp = new RegExp(responseHandlingConfig.code);
      return regExp.test(status.toString(10));
    }
    function resolveResponseHandling(xhr) {
      for (var i = 0; i < htmx.config.responseHandling.length; i++) {
        var responseHandlingElement = htmx.config.responseHandling[i];
        if (codeMatches(responseHandlingElement, xhr.status)) {
          return responseHandlingElement;
        }
      }
      return {
        swap: false
      };
    }
    function handleTitle(title) {
      if (title) {
        var titleElt = find("title");
        if (titleElt) {
          titleElt.innerHTML = title;
        } else {
          window.document.title = title;
        }
      }
    }
    function handleAjaxResponse(elt, responseInfo) {
      var xhr = responseInfo.xhr;
      var target = responseInfo.target;
      var etc = responseInfo.etc;
      var responseInfoSelect = responseInfo.select;
      if (!triggerEvent(elt, "htmx:beforeOnLoad", responseInfo)) return;
      if (hasHeader(xhr, /HX-Trigger:/i)) {
        handleTriggerHeader(xhr, "HX-Trigger", elt);
      }
      if (hasHeader(xhr, /HX-Location:/i)) {
        saveCurrentPageToHistory();
        var redirectPath = xhr.getResponseHeader("HX-Location");
        var redirectSwapSpec;
        if (redirectPath.indexOf("{") === 0) {
          redirectSwapSpec = parseJSON(redirectPath);
          redirectPath = redirectSwapSpec.path;
          delete redirectSwapSpec.path;
        }
        ajaxHelper("get", redirectPath, redirectSwapSpec).then(function() {
          pushUrlIntoHistory(redirectPath);
        });
        return;
      }
      var shouldRefresh = hasHeader(xhr, /HX-Refresh:/i) && xhr.getResponseHeader("HX-Refresh") === "true";
      if (hasHeader(xhr, /HX-Redirect:/i)) {
        responseInfo.keepIndicators = true;
        location.href = xhr.getResponseHeader("HX-Redirect");
        shouldRefresh && location.reload();
        return;
      }
      if (shouldRefresh) {
        responseInfo.keepIndicators = true;
        location.reload();
        return;
      }
      if (hasHeader(xhr, /HX-Retarget:/i)) {
        if (xhr.getResponseHeader("HX-Retarget") === "this") {
          responseInfo.target = elt;
        } else {
          responseInfo.target = asElement(querySelectorExt(elt, xhr.getResponseHeader("HX-Retarget")));
        }
      }
      var historyUpdate = determineHistoryUpdates(elt, responseInfo);
      var responseHandling = resolveResponseHandling(xhr);
      var shouldSwap = responseHandling.swap;
      var isError = !!responseHandling.error;
      var ignoreTitle = htmx.config.ignoreTitle || responseHandling.ignoreTitle;
      var selectOverride = responseHandling.select;
      if (responseHandling.target) {
        responseInfo.target = asElement(querySelectorExt(elt, responseHandling.target));
      }
      var swapOverride = etc.swapOverride;
      if (swapOverride == null && responseHandling.swapOverride) {
        swapOverride = responseHandling.swapOverride;
      }
      if (hasHeader(xhr, /HX-Retarget:/i)) {
        if (xhr.getResponseHeader("HX-Retarget") === "this") {
          responseInfo.target = elt;
        } else {
          responseInfo.target = asElement(querySelectorExt(elt, xhr.getResponseHeader("HX-Retarget")));
        }
      }
      if (hasHeader(xhr, /HX-Reswap:/i)) {
        swapOverride = xhr.getResponseHeader("HX-Reswap");
      }
      var serverResponse = xhr.response;
      var beforeSwapDetails = mergeObjects({
        shouldSwap: shouldSwap,
        serverResponse: serverResponse,
        isError: isError,
        ignoreTitle: ignoreTitle,
        selectOverride: selectOverride,
        swapOverride: swapOverride
      }, responseInfo);
      if (responseHandling.event && !triggerEvent(target, responseHandling.event, beforeSwapDetails)) return;
      if (!triggerEvent(target, "htmx:beforeSwap", beforeSwapDetails)) return;
      target = beforeSwapDetails.target;
      serverResponse = beforeSwapDetails.serverResponse;
      isError = beforeSwapDetails.isError;
      ignoreTitle = beforeSwapDetails.ignoreTitle;
      selectOverride = beforeSwapDetails.selectOverride;
      swapOverride = beforeSwapDetails.swapOverride;
      responseInfo.target = target;
      responseInfo.failed = isError;
      responseInfo.successful = !isError;
      if (beforeSwapDetails.shouldSwap) {
        if (xhr.status === 286) {
          cancelPolling(elt);
        }
        withExtensions(elt, function(extension) {
          serverResponse = extension.transformResponse(serverResponse, xhr, elt);
        });
        if (historyUpdate.type) {
          saveCurrentPageToHistory();
        }
        var swapSpec = getSwapSpecification(elt, swapOverride);
        if (!swapSpec.hasOwnProperty("ignoreTitle")) {
          swapSpec.ignoreTitle = ignoreTitle;
        }
        target.classList.add(htmx.config.swappingClass);
        var settleResolve = null;
        var settleReject = null;
        if (responseInfoSelect) {
          selectOverride = responseInfoSelect;
        }
        if (hasHeader(xhr, /HX-Reselect:/i)) {
          selectOverride = xhr.getResponseHeader("HX-Reselect");
        }
        var selectOOB = getClosestAttributeValue(elt, "hx-select-oob");
        var select = getClosestAttributeValue(elt, "hx-select");
        var doSwap = function doSwap2() {
          try {
            if (historyUpdate.type) {
              triggerEvent(getDocument().body, "htmx:beforeHistoryUpdate", mergeObjects({
                history: historyUpdate
              }, responseInfo));
              if (historyUpdate.type === "push") {
                pushUrlIntoHistory(historyUpdate.path);
                triggerEvent(getDocument().body, "htmx:pushedIntoHistory", {
                  path: historyUpdate.path
                });
              } else {
                replaceUrlInHistory(historyUpdate.path);
                triggerEvent(getDocument().body, "htmx:replacedInHistory", {
                  path: historyUpdate.path
                });
              }
            }
            swap(target, serverResponse, swapSpec, {
              select: selectOverride || select,
              selectOOB: selectOOB,
              eventInfo: responseInfo,
              anchor: responseInfo.pathInfo.anchor,
              contextElement: elt,
              afterSwapCallback: function afterSwapCallback() {
                if (hasHeader(xhr, /HX-Trigger-After-Swap:/i)) {
                  var finalElt = elt;
                  if (!bodyContains(elt)) {
                    finalElt = getDocument().body;
                  }
                  handleTriggerHeader(xhr, "HX-Trigger-After-Swap", finalElt);
                }
              },
              afterSettleCallback: function afterSettleCallback() {
                if (hasHeader(xhr, /HX-Trigger-After-Settle:/i)) {
                  var finalElt = elt;
                  if (!bodyContains(elt)) {
                    finalElt = getDocument().body;
                  }
                  handleTriggerHeader(xhr, "HX-Trigger-After-Settle", finalElt);
                }
                maybeCall(settleResolve);
              }
            });
          } catch (e) {
            triggerErrorEvent(elt, "htmx:swapError", responseInfo);
            maybeCall(settleReject);
            throw e;
          }
        };
        var shouldTransition = htmx.config.globalViewTransitions;
        if (swapSpec.hasOwnProperty("transition")) {
          shouldTransition = swapSpec.transition;
        }
        if (shouldTransition && triggerEvent(elt, "htmx:beforeTransition", responseInfo) && typeof Promise !== "undefined" && // @ts-ignore experimental feature atm
        document.startViewTransition) {
          var settlePromise = new Promise(function(_resolve, _reject) {
            settleResolve = _resolve;
            settleReject = _reject;
          });
          var innerDoSwap = doSwap;
          doSwap = function doSwap2() {
            document.startViewTransition(function() {
              innerDoSwap();
              return settlePromise;
            });
          };
        }
        if (swapSpec.swapDelay > 0) {
          getWindow().setTimeout(doSwap, swapSpec.swapDelay);
        } else {
          doSwap();
        }
      }
      if (isError) {
        triggerErrorEvent(elt, "htmx:responseError", mergeObjects({
          error: "Response Status Error Code " + xhr.status + " from " + responseInfo.pathInfo.requestPath
        }, responseInfo));
      }
    }
    var extensions = {};
    function extensionBase() {
      return {
        init: function init2(api) {
          return null;
        },
        getSelectors: function getSelectors() {
          return null;
        },
        onEvent: function onEvent(name, evt) {
          return true;
        },
        transformResponse: function transformResponse(text, xhr, elt) {
          return text;
        },
        isInlineSwap: function isInlineSwap2(swapStyle) {
          return false;
        },
        handleSwap: function handleSwap(swapStyle, target, fragment, settleInfo) {
          return false;
        },
        encodeParameters: function encodeParameters(xhr, parameters, elt) {
          return null;
        }
      };
    }
    function defineExtension(name, extension) {
      if (extension.init) {
        extension.init(internalAPI);
      }
      extensions[name] = mergeObjects(extensionBase(), extension);
    }
    function removeExtension(name) {
      delete extensions[name];
    }
    function getExtensions(elt, extensionsToReturn, extensionsToIgnore) {
      if (extensionsToReturn == void 0) {
        extensionsToReturn = [];
      }
      if (elt == void 0) {
        return extensionsToReturn;
      }
      if (extensionsToIgnore == void 0) {
        extensionsToIgnore = [];
      }
      var extensionsForElement = getAttributeValue(elt, "hx-ext");
      if (extensionsForElement) {
        forEach(extensionsForElement.split(","), function(extensionName) {
          extensionName = extensionName.replace(/ /g, "");
          if (extensionName.slice(0, 7) == "ignore:") {
            extensionsToIgnore.push(extensionName.slice(7));
            return;
          }
          if (extensionsToIgnore.indexOf(extensionName) < 0) {
            var extension = extensions[extensionName];
            if (extension && extensionsToReturn.indexOf(extension) < 0) {
              extensionsToReturn.push(extension);
            }
          }
        });
      }
      return getExtensions(asElement(parentElt(elt)), extensionsToReturn, extensionsToIgnore);
    }
    var isReady = false;
    getDocument().addEventListener("DOMContentLoaded", function() {
      isReady = true;
    });
    function ready(fn) {
      if (isReady || getDocument().readyState === "complete") {
        fn();
      } else {
        getDocument().addEventListener("DOMContentLoaded", fn);
      }
    }
    function insertIndicatorStyles() {
      if (htmx.config.includeIndicatorStyles !== false) {
        var nonceAttribute = htmx.config.inlineStyleNonce ? ' nonce="'.concat(htmx.config.inlineStyleNonce, '"') : "";
        getDocument().head.insertAdjacentHTML("beforeend", "<style" + nonceAttribute + ">      ." + htmx.config.indicatorClass + "{opacity:0}      ." + htmx.config.requestClass + " ." + htmx.config.indicatorClass + "{opacity:1; transition: opacity 200ms ease-in;}      ." + htmx.config.requestClass + "." + htmx.config.indicatorClass + "{opacity:1; transition: opacity 200ms ease-in;}      </style>");
      }
    }
    function getMetaConfig() {
      var element = getDocument().querySelector('meta[name="htmx-config"]');
      if (element) {
        return parseJSON(element.content);
      } else {
        return null;
      }
    }
    function mergeMetaConfig() {
      var metaConfig = getMetaConfig();
      if (metaConfig) {
        htmx.config = mergeObjects(htmx.config, metaConfig);
      }
    }
    ready(function() {
      mergeMetaConfig();
      insertIndicatorStyles();
      var body = getDocument().body;
      processNode(body);
      var restoredElts = getDocument().querySelectorAll("[hx-trigger='restored'],[data-hx-trigger='restored']");
      body.addEventListener("htmx:abort", function(evt) {
        var target = evt.target;
        var internalData = getInternalData(target);
        if (internalData && internalData.xhr) {
          internalData.xhr.abort();
        }
      });
      var originalPopstate = window.onpopstate ? window.onpopstate.bind(window) : null;
      window.onpopstate = function(event2) {
        if (event2.state && event2.state.htmx) {
          restoreHistory();
          forEach(restoredElts, function(elt) {
            triggerEvent(elt, "htmx:restored", {
              document: getDocument(),
              triggerEvent: triggerEvent
            });
          });
        } else {
          if (originalPopstate) {
            originalPopstate(event2);
          }
        }
      };
      getWindow().setTimeout(function() {
        triggerEvent(body, "htmx:load", {});
        body = null;
      }, 0);
    });
    return htmx;
  }();
  var htmx_esm_default = htmx2;
  var isFullscreen = false;
  var fullscreenAPI = getFullscreenAPI();
  function getFullscreenAPI() {
    var apis = [["requestFullscreen", "exitFullscreen", "fullscreenElement", "fullscreenEnabled"], ["mozRequestFullScreen", "mozCancelFullScreen", "mozFullScreenElement", "mozFullScreenEnabled"], ["webkitRequestFullscreen", "webkitExitFullscreen", "webkitFullscreenElement", "webkitFullscreenEnabled"], ["msRequestFullscreen", "msExitFullscreen", "msFullscreenElement", "msFullscreenEnabled"]];
    for (var _i2 = 0, _apis = apis; _i2 < _apis.length; _i2++) {
      var _apis$_i = _slicedToArray(_apis[_i2], 4), request = _apis$_i[0], exit = _apis$_i[1], element = _apis$_i[2], enabled = _apis$_i[3];
      if (request in document.documentElement) {
        return {
          requestFullscreen: request,
          exitFullscreen: exit,
          fullscreenElement: element,
          fullscreenEnabled: enabled
        };
      }
    }
    return {
      requestFullscreen: null,
      exitFullscreen: null,
      fullscreenElement: null,
      fullscreenEnabled: null
    };
  }
  function toggleFullscreen(documentBody2, fullscreenButton2) {
    var _a2;
    if (isFullscreen) {
      if (fullscreenAPI.exitFullscreen) {
        document[fullscreenAPI.exitFullscreen]();
      }
    } else {
      (_a2 = documentBody2[fullscreenAPI.requestFullscreen]) == null ? void 0 : _a2.call(documentBody2);
    }
    isFullscreen = !isFullscreen;
    fullscreenButton2 == null ? void 0 : fullscreenButton2.classList.toggle("navigation--fullscreen-enabled");
  }
  function addFullscreenEventListener(fullscreenButton2) {
    document.addEventListener("fullscreenchange", function() {
      isFullscreen = !!document[fullscreenAPI.fullscreenElement];
      fullscreenButton2 == null ? void 0 : fullscreenButton2.classList.toggle("navigation--fullscreen-enabled", isFullscreen);
    });
  }
  var animationFrameId = null;
  var progressBarElement;
  var lastPollTime = null;
  var pausedTime = null;
  var isPaused = false;
  var pollInterval;
  var kioskElement;
  var menuElement;
  var menuPausePlayButton;
  function initPolling(interval, kiosk2, menu2, pausePlayButton) {
    pollInterval = interval;
    kioskElement = kiosk2;
    menuElement = menu2;
    menuPausePlayButton = pausePlayButton;
  }
  function updateKiosk(timestamp) {
    if (pausedTime !== null) {
      lastPollTime += timestamp - pausedTime;
      pausedTime = null;
    }
    var elapsed = timestamp - lastPollTime;
    var progress = Math.min(elapsed / pollInterval, 1);
    if (progressBarElement) {
      progressBarElement.style.width = "".concat(progress * 100, "%");
    }
    if (elapsed >= pollInterval) {
      htmx_esm_default.trigger(kioskElement, "kiosk-new-image");
      lastPollTime = timestamp;
      stopPolling();
      return;
    }
    animationFrameId = requestAnimationFrame(updateKiosk);
  }
  function _startPolling() {
    progressBarElement = htmx_esm_default.find(".progress--bar");
    progressBarElement == null ? void 0 : progressBarElement.classList.remove("progress--bar-paused");
    menuPausePlayButton == null ? void 0 : menuPausePlayButton.classList.remove("navigation--control--paused");
    lastPollTime = performance.now();
    pausedTime = null;
    animationFrameId = requestAnimationFrame(updateKiosk);
  }
  function stopPolling() {
    if (isPaused && animationFrameId === null) return;
    cancelAnimationFrame(animationFrameId);
    progressBarElement == null ? void 0 : progressBarElement.classList.add("progress--bar-paused");
    menuPausePlayButton == null ? void 0 : menuPausePlayButton.classList.add("navigation--control--paused");
  }
  function pausePolling() {
    if (isPaused && animationFrameId === null) return;
    cancelAnimationFrame(animationFrameId);
    pausedTime = performance.now();
    progressBarElement == null ? void 0 : progressBarElement.classList.add("progress--bar-paused");
    menuPausePlayButton == null ? void 0 : menuPausePlayButton.classList.add("navigation--control--paused");
    menuElement == null ? void 0 : menuElement.classList.remove("navigation-hidden");
    isPaused = true;
  }
  function resumePolling() {
    if (!isPaused) return;
    animationFrameId = requestAnimationFrame(updateKiosk);
    progressBarElement == null ? void 0 : progressBarElement.classList.remove("progress--bar-paused");
    menuPausePlayButton == null ? void 0 : menuPausePlayButton.classList.remove("navigation--control--paused");
    menuElement == null ? void 0 : menuElement.classList.add("navigation-hidden");
    isPaused = false;
  }
  function togglePolling() {
    isPaused ? resumePolling() : pausePolling();
  }
  var preventSleep = function preventSleep2() {
    return __async(void 0, null, /* @__PURE__ */ _regeneratorRuntime().mark(function _callee3() {
      var wakeLock, requestWakeLock, handleVisibilityChange;
      return _regeneratorRuntime().wrap(function _callee3$(_context3) {
        while (1) switch (_context3.prev = _context3.next) {
          case 0:
            wakeLock = null;
            requestWakeLock = function requestWakeLock2() {
              return __async(void 0, null, /* @__PURE__ */ _regeneratorRuntime().mark(function _callee() {
                return _regeneratorRuntime().wrap(function _callee$(_context) {
                  while (1) switch (_context.prev = _context.next) {
                    case 0:
                      if (!("wakeLock" in navigator)) {
                        _context.next = 24;
                        break;
                      }
                      _context.prev = 1;
                      _context.next = 4;
                      return navigator.wakeLock.request("screen");
                    case 4:
                      wakeLock = _context.sent;
                      wakeLock.addEventListener("release", function() {
                        wakeLock = null;
                      });
                      _context.next = 24;
                      break;
                    case 8:
                      _context.prev = 8;
                      _context.t0 = _context["catch"](1);
                      if (!(_context.t0 instanceof TypeError)) {
                        _context.next = 23;
                        break;
                      }
                      _context.prev = 11;
                      _context.next = 14;
                      return navigator.wakeLock.request();
                    case 14:
                      wakeLock = _context.sent;
                      wakeLock.addEventListener("release", function() {
                        wakeLock = null;
                      });
                      _context.next = 21;
                      break;
                    case 18:
                      _context.prev = 18;
                      _context.t1 = _context["catch"](11);
                      console.error("Failed to acquire Wake Lock:", _context.t1);
                    case 21:
                      _context.next = 24;
                      break;
                    case 23:
                      console.error("Error acquiring Wake Lock:", _context.t0);
                    case 24:
                    case "end":
                      return _context.stop();
                  }
                }, _callee, null, [[1, 8], [11, 18]]);
              }));
            };
            handleVisibilityChange = function handleVisibilityChange2() {
              return __async(void 0, null, /* @__PURE__ */ _regeneratorRuntime().mark(function _callee2() {
                return _regeneratorRuntime().wrap(function _callee2$(_context2) {
                  while (1) switch (_context2.prev = _context2.next) {
                    case 0:
                      if (!(document.visibilityState === "visible")) {
                        _context2.next = 3;
                        break;
                      }
                      _context2.next = 3;
                      return requestWakeLock();
                    case 3:
                    case "end":
                      return _context2.stop();
                  }
                }, _callee2);
              }));
            };
            document.addEventListener("visibilitychange", handleVisibilityChange);
            _context3.next = 6;
            return requestWakeLock();
          case 6:
            return _context3.abrupt("return", function() {
              document.removeEventListener("visibilitychange", handleVisibilityChange);
              wakeLock == null ? void 0 : wakeLock.release();
            });
          case 7:
          case "end":
            return _context3.stop();
        }
      }, _callee3);
    }));
  };
  var _a;
  var kioskData = JSON.parse(((_a = document.getElementById("kiosk-data")) == null ? void 0 : _a.textContent) || "{}");
  var pollInterval2 = htmx_esm_default.parseInterval("".concat(kioskData.refresh, "s"));
  var documentBody = document.body;
  var fullscreenButton = htmx_esm_default.find(".navigation--fullscreen");
  var fullScreenButtonSeperator = htmx_esm_default.find(".navigation--fullscreen-separator");
  var kiosk = htmx_esm_default.find("#kiosk");
  var menu = htmx_esm_default.find(".navigation");
  var menuInteraction = htmx_esm_default.find("#navigation-interaction-area");
  var menuPausePlayButton2 = htmx_esm_default.find(".navigation--control");
  function init() {
    return __async(this, null, /* @__PURE__ */ _regeneratorRuntime().mark(function _callee4() {
      return _regeneratorRuntime().wrap(function _callee4$(_context4) {
        while (1) switch (_context4.prev = _context4.next) {
          case 0:
            if (kioskData.debugVerbose) {
              htmx_esm_default.logAll();
            }
            if (!kioskData.disableScreensaver) {
              _context4.next = 4;
              break;
            }
            _context4.next = 4;
            return preventSleep();
          case 4:
            if ("serviceWorker" in navigator) {
              navigator.serviceWorker.register("/assets/js/sw.js").then(function(registration) {
                console.log("ServiceWorker registration successful");
              }, function(err) {
                console.log("ServiceWorker registration failed: ", err);
              });
            }
            if (!fullscreenAPI.requestFullscreen) {
              fullscreenButton && htmx_esm_default.remove(fullscreenButton);
              fullScreenButtonSeperator && htmx_esm_default.remove(fullScreenButtonSeperator);
            }
            if (pollInterval2) {
              initPolling(pollInterval2, kiosk, menu, menuPausePlayButton2);
            } else {
              console.error("Could not start polling");
            }
            addEventListeners();
          case 8:
          case "end":
            return _context4.stop();
        }
      }, _callee4);
    }));
  }
  function handleFullscreenClick() {
    toggleFullscreen(documentBody, fullscreenButton);
  }
  function addEventListeners() {
    menuInteraction == null ? void 0 : menuInteraction.addEventListener("click", togglePolling);
    menuPausePlayButton2 == null ? void 0 : menuPausePlayButton2.addEventListener("click", togglePolling);
    fullscreenButton == null ? void 0 : fullscreenButton.addEventListener("click", handleFullscreenClick);
    addFullscreenEventListener(fullscreenButton);
    htmx_esm_default.on("htmx:afterRequest", function(e) {
      var offlineSVG = htmx_esm_default.find("#offline");
      if (!offlineSVG) {
        console.error("offline svg missing");
        return;
      }
      if (e.detail.successful) {
        htmx_esm_default.removeClass(offlineSVG, "offline");
      } else {
        htmx_esm_default.addClass(offlineSVG, "offline");
      }
    });
  }
  function _cleanupFrames() {
    var frames = htmx_esm_default.findAll(".frame");
    if (frames.length > 3) {
      htmx_esm_default.remove(frames[0], 3e3);
    }
  }
  document.addEventListener("DOMContentLoaded", function() {
    init();
  });
  return __toCommonJS(kiosk_exports);
}();
/*! regenerator-runtime -- Copyright (c) 2014-present, Facebook, Inc. -- license (MIT): https://github.com/facebook/regenerator/blob/main/LICENSE */
