import babel from '@rollup/plugin-babel'
import postcss from 'rollup-plugin-postcss'
const html  = require('@rollup/plugin-html')
import image from '@rollup/plugin-image';
import json from '@rollup/plugin-json'
import { nodeResolve } from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import { terser } from "rollup-plugin-terser";

export default {
  input: 'src/main.js',
  output: [
    { file: "dist/bundle.js", format: "cjs" },
    { file: "dist/bundle.min.js", format: "cjs", plugins: [terser()] },
    { file: "dist/bundle.esm.js", format: "esm" },
  ],
  plugins: [
     nodeResolve(), 
     postcss({
        extract: true,
        extract: 'bundle.css',
        plugins: []
      }), 
     babel({ 
      exclude: 'node_modules/**',
      babelHelpers: 'bundled' 
     }), 
     commonjs(), 
     html({
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'minimum-scale=1, initial-scale=1, width=device-width' },
        { name: 'apple-mobile-web-app-status-bar-style', content: 'black-translucent' },
        { name: 'apple-mobile-web-app-capable', content: 'yes' },
        { name: 'mobile-web-app-capable', content: 'yes' },
        { name: 'HandheldFriendly', content: 'True' },
        { name: 'MobileOptimized', content: '320' },
        { name: 'robots', content: 'noindex,nofollow,noarchive' },
        { name: 'X-UA-Compatible', content: 'ie=edge' },
        { name: 'description', content: '' },
        { name: 'title', content: '' }
      ]
     }), image(), json()],
}
