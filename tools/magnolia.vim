" place this in the init path (.vimrc)
" autocmd BufNewFile,BufRead *.mgn set filetype=magnolia

if exists("b:current_syntax")
    finish
endif

" mgn syntax definition for vi/vim
syntax sync fromstart

" prefer hard tabs
set noexpandtab

" case
syntax match mgnLabel "\v\:"
syntax match mgnLabel "\v\-\>"
highlight link mgnCase Label

" operators
syntax match mgnOp "\v\~"
syntax match mgnOp "\v\+"
syntax match mgnOp "\v\-"
syntax match mgnOp "\v\*"
syntax match mgnOp "\v\/"
syntax match mgnOp "\v\%"

syntax match mgnOp "\v\&"
syntax match mgnOp "\v\|"
syntax match mgnOp "\v\^"

syntax match mgnOp "\v\>"
syntax match mgnOp "\v\<"
syntax match mgnOp "\v\="
syntax match mgnOp "\v\>\="
syntax match mgnOp "\v\<\="
syntax match mgnOp "\v\."
syntax match mgnOp "\v\:\="
syntax match mgnOp "\v\<\-"
highlight link mgnOp Operator

" match
syntax keyword mgnMatch if
syntax match mgnMatch "\v\-\>"
highlight link mgnMatch Conditional

" functions
syntax keyword mgnFunction fn
syntax keyword mgnFunction with
syntax match mgnFunction "\v\|\>"
highlight link mgnFunction Type

" booleans
syntax keyword mgnBoolean true false
highlight link mgnBoolean Boolean

" constants
syntax keyword mgnConst _
highlight link mgnConst Constant

" atoms
syntax match mgnAtom "\v:[A-Za-z_!][A-Za-z0-9_!?]*"
highlight link mgnAtom Special

" numbers should be consumed first by identifiers, so comes before
syntax match mgnNumber "\v\d+"
syntax match mgnNumber "\v\d+\.\d+"
highlight link mgnNumber Number

" functions
syntax match mgnFnCall "\v[A-Za-z_!][A-Za-z0-9_!?]*\(" contains=mgnFunctionName,mgnBuiltin

" identifiers
syntax match mgnFunctionName "\v[A-Za-z_!][A-Za-z0-9_!?]*" contained
highlight link mgnFunctionName Identifier

syntax keyword mgnBuiltin import contained
syntax keyword mgnBuiltin string contained
syntax keyword mgnBuiltin int contained
syntax keyword mgnBuiltin float contained
syntax keyword mgnBuiltin atom contained
syntax keyword mgnBuiltin codepoint contained
syntax keyword mgnBuiltin char contained
syntax keyword mgnBuiltin type contained
syntax keyword mgnBuiltin len contained
syntax keyword mgnBuiltin keys contained

syntax keyword mgnBuiltin args contained
syntax keyword mgnBuiltin env contained
syntax keyword mgnBuiltin time contained
syntax keyword mgnBuiltin exit contained
syntax keyword mgnBuiltin rand contained

syntax keyword mgnBuiltin wait contained
syntax keyword mgnBuiltin exec contained
syntax keyword mgnBuiltin input contained
syntax keyword mgnBuiltin print contained
syntax keyword mgnBuiltin sleep contained
syntax keyword mgnBuiltin ls contained
syntax keyword mgnBuiltin mkdir contained
syntax keyword mgnBuiltin rm contained
syntax keyword mgnBuiltin stat contained
syntax keyword mgnBuiltin open contained
syntax keyword mgnBuiltin read contained
syntax keyword mgnBuiltin write contained
syntax keyword mgnBuiltin close contained
syntax keyword mgnBuiltin listen contained
syntax keyword mgnBuiltin req contained
syntax keyword mgnBuiltin sin contained
syntax keyword mgnBuiltin cos contained
syntax keyword mgnBuiltin tan contained
syntax keyword mgnBuiltin asin contained
syntax keyword mgnBuiltin acos contained
syntax keyword mgnBuiltin atan contained
syntax keyword mgnBuiltin pow contained
syntax keyword mgnBuiltin ln contained
highlight link mgnBuiltin Keyword

" strings
syntax region mgnString start=/\v'/ skip=/\v(\\.|\r|\n)/ end=/\v'/
highlight link mgnString String

" comment
" -- block
syntax region mgnComment start=/\v\/\*/ skip=/\v(\\.|\r|\n)/ end=/\v\*\// contains=mgnTodo
highlight link mgnComment Comment
" -- line-ending comment
syntax match mgnLineComment "\v\/\/.*" contains=mgnTodo
highlight link mgnLineComment Comment
" -- shebang, highlighted as comment
syntax match mgnShebangComment "\v^#!.*"
highlight link mgnShebangComment Comment
" -- TODO in comments
syntax match mgnTodo "\v(TODO\(.*\)|TODO)" contained
syntax keyword mgnTodo XXX contained
highlight link mgnTodo Todo

" syntax-based code folds
syntax region mgnExpressionList start="(" end=")" transparent fold
syntax region mgnMatchExpression start="{" end="}" transparent fold
syntax region mgnComposite start="\v\[" end="\v\]" transparent fold
set foldmethod=syntax

let b:current_syntax = "mgn"
