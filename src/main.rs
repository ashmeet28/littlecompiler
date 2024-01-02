use std::env;
use std::fs;

fn generate_tokens(source_code_bytes: Vec<u8>) {
    dbg!(source_code_bytes);
}
fn main() {
    let args: Vec<String> = env::args().collect();
    let file_path = args[1].clone();
    let file_contents = fs::read(file_path).unwrap();
    generate_tokens(file_contents);
}
