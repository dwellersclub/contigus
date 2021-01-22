
import resolve from '@rollup/plugin-node-resolve';
import postcss from 'rollup-plugin-postcss'
import path from 'path';
import babel from '@rollup/plugin-babel'
import copy from 'rollup-plugin-copy-watch';
import minifyHTML from 'rollup-plugin-minify-html-literals';
import cssnano from 'cssnano';
import cssImport from 'postcss-import';
import purgecss from '@fullhuman/postcss-purgecss';

import image from '@rollup/plugin-image';
import json from '@rollup/plugin-json'
import injectProcessEnv from 'rollup-plugin-inject-process-env';
import commonjs from '@rollup/plugin-commonjs';
import { terser } from "rollup-plugin-terser";

const production = !process.env.ROLLUP_WATCH;

export default {
  input: ["src/index.js", "src/apps/inventory/inventory.js"],
  output: { dir: 'dist', format: 'es', sourcemap: true },
  plugins: [
    minifyHTML(),
    copy({
      watch: production ? null : ['src/**/*.html'],
      targets: [
        { src: 'node_modules/@cds/city/Webfonts', dest: 'dist' },
        { src: 'src/index.html', dest: 'dist' },
      ],
    }), babel({ 
      exclude: 'node_modules/**',
      babelHelpers: 'bundled',
     }), 
    postcss({
      extract: true,
      extract: path.resolve('dist/index.css'),
      plugins: [
        cssImport(),
        production &&
          purgecss({
            content: ['./src/**/*.html'],
            variables: true,
            // custom matcher to better find Clarity utilities with cds-text and cds-layout
            defaultExtractor: content => content.match(/[\w-\/:@]+(?<!:)/g) || [],
          }),
        cssnano(),
      ],
    }),
     resolve(), 
     production && terser({ output: { comments: false } }),
     commonjs(), 
     injectProcessEnv({ 
          NODE_ENV: 'production'
      }), image(), json()],
}
