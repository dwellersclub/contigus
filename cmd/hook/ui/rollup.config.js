import babel from '@rollup/plugin-babel'
import sass from 'rollup-plugin-sass'
const html  = require('@rollup/plugin-html')
import image from '@rollup/plugin-image';
import json from '@rollup/plugin-json'
import { nodeResolve } from '@rollup/plugin-node-resolve';
import commonjs from '@rollup/plugin-commonjs';
import { terser } from "rollup-plugin-terser";
import pkg from './package.json'

export default {
  input: 'src/main.js',
  output: [
    { file: "dist/bundle.js", format: "cjs" },
    { file: "dist/bundle.min.js", format: "cjs", plugins: [terser()] },
    { file: "dist/bundle.esm.js", format: "esm" },
  ],
  plugins: [
     nodeResolve(), 
     babel({ 
      exclude: 'node_modules/**',
      babelHelpers: 'bundled' 
     }), 
     commonjs(), 
     sass(), 
     html(), image(), json()],
}
