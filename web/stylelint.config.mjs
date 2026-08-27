export default {
  extends: [
    'stylelint-config-standard',
    'stylelint-config-standard-scss',
    'stylelint-config-recommended-vue/scss',
  ],
  overrides: [
    {
      files: ['**/*.vue'],
      customSyntax: 'postcss-html',
      rules: {
        'no-invalid-position-declaration': null,
      },
    },
    { files: ['**/*.scss'], customSyntax: 'postcss-scss' },
  ],
  rules: {
    'no-descending-specificity': null,
    'no-empty-source': null,
    'selector-class-pattern': null,
    'declaration-empty-line-before': null,
    'custom-property-empty-line-before': null,
    'value-keyword-case': null,
    'selector-pseudo-element-no-unknown': [
      true,
      { ignorePseudoElements: ['v-deep', 'v-slotted', 'v-global'] },
    ],
    'at-rule-no-unknown': [
      true,
      {
        ignoreAtRules: [
          'apply',
          'screen',
          'variants',
          'tailwind',
          'use',
          'forward',
          'mixin',
          'include',
          'if',
          'else',
          'each',
          'for',
          'function',
          'return',
        ],
      },
    ],
  },
  ignoreFiles: ['dist/**', 'node_modules/**', 'public/**', 'src/**/*.d.ts'],
}
