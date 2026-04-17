module.exports = {
  "disallow-foreach": {
    meta: { type: "suggestion", docs: { description: "Disallow Array.forEach in favor of for loops" }, schema: [] },
    create(context) {
      return {
        CallExpression(node) {
          if (node.callee.type === "MemberExpression" && node.callee.property.name === "forEach") {
            context.report({ node, message: "Use 'for (...)' instead of '.forEach()'" });
          }
        }
      };
    }
  },
  "disallow-then": {
    meta: { type: "suggestion", docs: { description: "Disallow .then() in favor of async/await" }, schema: [] },
    create(context) {
      return {
        CallExpression(node) {
          if (node.callee.type === "MemberExpression" && node.callee.property.name === "then") {
            context.report({ node, message: "Use 'await' instead of '.then()'" });
          }
        }
      };
    }
  },
  "disallow-catch": {
    meta: { type: "suggestion", docs: { description: "Disallow .catch() in favor of try/catch with await" }, schema: [] },
    create(context) {
      return {
        CallExpression(node) {
          if (node.callee.type === "MemberExpression" && node.callee.property.name === "catch" &&
              node.callee.object.type === "CallExpression") {
            context.report({ node, message: "Use try/catch with 'await' instead of '.catch()'" });
          }
        }
      };
    }
  },
  "disallow-axios": {
    meta: { type: "suggestion", docs: { description: "Disallow axios imports" }, schema: [] },
    create(context) {
      return {
        ImportDeclaration(node) {
          if (node.source.value === "axios") {
            context.report({ node, message: "Use the API client from '@/lib/api' instead of axios" });
          }
        }
      };
    }
  },
  "disallow-fetch": {
    meta: { type: "suggestion", docs: { description: "Disallow raw fetch() calls" }, schema: [] },
    create(context) {
      return {
        CallExpression(node) {
          if (node.callee.name === "fetch") {
            const sourceCode = context.getSourceCode();
            const comments = sourceCode.getCommentsBefore(node);
            const hasAllowComment = comments.some(c => c.value.includes("eslint-allow-fetch"));
            if (!hasAllowComment) {
              context.report({ node, message: "Use the API client from '@/lib/api' instead of raw fetch()" });
            }
          }
        }
      };
    }
  },
  "disallow-multiple-function-arguments": {
    meta: { type: "suggestion", docs: { description: "Limit function parameters to 2" }, schema: [] },
    create(context) {
      return {
        FunctionDeclaration(node) {
          if (node.params.length > 2) {
            context.report({ node, message: `Function has ${node.params.length} parameters. Use an object argument when more than 2 parameters are needed.` });
          }
        },
        ArrowFunctionExpression(node) {
          if (node.params.length > 2) {
            context.report({ node, message: `Arrow function has ${node.params.length} parameters. Use an object argument when more than 2 parameters are needed.` });
          }
        }
      };
    }
  },
  "enforce-readonly-fields": {
    meta: { type: "suggestion", docs: { description: "Enforce readonly on TypeScript interface/type fields" }, fixable: "code", schema: [] },
    create(context) {
      return {
        TSPropertySignature(node) {
          if (!node.readonly) {
            context.report({
              node,
              message: "Interface/type fields should be readonly",
              fix(fixer) {
                return fixer.insertTextBefore(node, "readonly ");
              }
            });
          }
        }
      };
    }
  },
  "enforce-readonly-array": {
    meta: { type: "suggestion", docs: { description: "Enforce readonly arrays in TypeScript" }, fixable: "code", schema: [] },
    create(context) {
      return {
        TSArrayType(node) {
          if (node.parent.type !== "TSTypeOperator" || node.parent.operator !== "readonly") {
            context.report({
              node,
              message: "Use 'readonly T[]' instead of 'T[]'"
            });
          }
        }
      };
    }
  },
  "disallow-empty-string": {
    meta: { type: "suggestion", docs: { description: "Disallow empty string literals" }, schema: [] },
    create(context) {
      return {
        Literal(node) {
          if (node.value === "" && node.parent.type !== "BinaryExpression") {
            context.report({ node, message: "Prefer null or undefined over empty string" });
          }
        }
      };
    }
  }
};
