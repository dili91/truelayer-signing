# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## 0.1.7
* When verifying permit signed/verified path single trailing slash mismatches.

## 0.1.6
* Added `Jku` method to `Signer` to allow different jku's endpoints.

## 0.1.5
* Add `path` arg validation to `Signer` & `Verifier` for more informative errors.
* Fixed README example
* Removed license symlinks

## 0.1.4
* Added `Headers` method to `Signer` and `Verifier`
* Updated README

## 0.1.3
* Added support for verifying jwks with alg: `ES512`.
* Added verifying webhooks readme examples.

## 0.1.2
* Added `VerifyWithJwks` method to aid verifying webhook signatures.

## 0.1.1
* Removed `Verifier`'s `method` field default value

## 0.1.0
* Added `SignWithPem`, `VerifyWithPem` and `ExtractJwsHeader`.
