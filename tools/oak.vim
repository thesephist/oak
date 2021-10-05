" place this in the init path (.vimrc)
" autocmd BufNewFile,BufRead *.oak set filetype=oak

if exists("b:current_syntax")
    finish
endif

" oak syntax definition for vi/vim
syntax sync fromstart

" prefer hard tabs
set noexpandtab

" operators
syntax match oakOp "\v\~"
syntax match oakOp "\v\+"
syntax match oakOp "\v\-"
syntax match oakOp "\v\*"
syntax match oakOp "\v\/"
syntax match oakOp "\v\%"

syntax match oakOp "\v\&"
syntax match oakOp "\v\|"
syntax match oakOp "\v\^"

syntax match oakOp "\v\>"
syntax match oakOp "\v\<"
syntax match oakOp "\v\="
syntax match oakOp "\v\>\="
syntax match oakOp "\v\<\="
syntax match oakOp "\v\!\="
syntax match oakOp "\v\."
syntax match oakOp "\v\:\="
syntax match oakOp "\v\<\-"
syntax match oakOp "\v\<\<"
highlight link oakOp Operator

" match
syntax keyword oakMatch if
syntax match oakMatch "\v\-\>"
highlight link oakMatch Conditional

" functions
syntax keyword oakFunction fn
syntax keyword oakFunction with
syntax match oakFunction "\v\|\>"
highlight link oakFunction Type

" bools
syntax keyword oakBool true false
highlight link oakBool Boolean

" constants
syntax keyword oakConst _
highlight link oakConst Constant

" atoms
syntax match oakAtom "\v:[A-Za-z_!][A-Za-z0-9_!?]*"
highlight link oakAtom Special

" numbers should be consumed first by identifiers, so comes before
syntax match oakNumber "\v\d+"
syntax match oakNumber "\v\d+\.\d+"
highlight link oakNumber Number

" functions
syntax match oakFnCall "\v[A-Za-z_!][A-Za-z0-9_!?]*\(" contains=oakFunctionName,oakBuiltin

" identifiers
syntax match oakFunctionName "\v[A-Za-z_][A-Za-z0-9_!?]*" contained
highlight link oakFunctionName Identifier

syntax keyword oakBuiltin import contained
syntax keyword oakBuiltin string contained
syntax keyword oakBuiltin int contained
syntax keyword oakBuiltin float contained
syntax keyword oakBuiltin atom contained
syntax keyword oakBuiltin codepoint contained
syntax keyword oakBuiltin char contained
syntax keyword oakBuiltin type contained
syntax keyword oakBuiltin len contained
syntax keyword oakBuiltin keys contained

syntax keyword oakBuiltin args contained
syntax keyword oakBuiltin env contained
syntax keyword oakBuiltin time contained
syntax keyword oakBuiltin nanotime contained
syntax keyword oakBuiltin exit contained
syntax keyword oakBuiltin rand contained
syntax keyword oakBuiltin wait contained
syntax keyword oakBuiltin exec contained

syntax keyword oakBuiltin input contained
syntax keyword oakBuiltin print contained
syntax keyword oakBuiltin ls contained
syntax keyword oakBuiltin mkdir contained
syntax keyword oakBuiltin rm contained
syntax keyword oakBuiltin stat contained
syntax keyword oakBuiltin open contained
syntax keyword oakBuiltin close contained
syntax keyword oakBuiltin read contained
syntax keyword oakBuiltin write contained
syntax keyword oakBuiltin listen contained
syntax keyword oakBuiltin req contained

syntax keyword oakBuiltin sin contained
syntax keyword oakBuiltin cos contained
syntax keyword oakBuiltin tan contained
syntax keyword oakBuiltin asin contained
syntax keyword oakBuiltin acos contained
syntax keyword oakBuiltin atan contained
syntax keyword oakBuiltin pow contained
syntax keyword oakBuiltin log contained
highlight link oakBuiltin Keyword

" strings
syntax region oakString start=/\v'/ skip=/\v(\\.|\r|\n)/ end=/\v'/
highlight link oakString String

" comment
" -- block
syntax region oakComment start=/\v\/\*/ skip=/\v(\\.|\r|\n)/ end=/\v\*\// contains=oakTodo
highlight link oakComment Comment
" -- line-ending comment
syntax match oakLineComment "\v\/\/.*" contains=oakTodo
highlight link oakLineComment Comment
" -- shebang, highlighted as comment
syntax match oakShebangComment "\v^#!.*"
highlight link oakShebangComment Comment
" -- TODO in comments
syntax match oakTodo "\v(TODO\(.*\)|TODO)" contained
syntax match oakTodo "\v(NOTE\(.*\)|NOTE)" contained
syntax match oakTodo "\v(XXX\(.*\)|XXX)" contained
syntax keyword oakTodo XXX contained
highlight link oakTodo Todo

" syntax-based code folds
syntax region oakExpressionList start="(" end=")" transparent fold
syntax region oakMatchExpression start="{" end="}" transparent fold
syntax region oakComposite start="\v\[" end="\v\]" transparent fold
set foldmethod=syntax

let b:current_syntax = "oak"
