use futures::{future::join_all, TryFutureExt};

#[tokio::main]
async fn main() -> anyhow::Result<()> {
    env_logger::init();

    let client = reqwest::Client::builder()
        .danger_accept_invalid_certs(true)
        .build()?;

    let futs = (0..get_concurrency())
        .map(|i| request(client.clone(), i).inspect_err(|e| eprintln!("❌ {e:?}")));
    join_all(futs).await;
    println!("done");
    Ok(())
}

fn get_concurrency() -> i32 {
    let Ok(x) = std::env::var("CONCURRENCY") else {
        return 1;
    };
    let Ok(v) = x.parse() else {
        return 1;
    };
    v
}

async fn request(client: reqwest::Client, id: i32) -> anyhow::Result<()> {
    let url = "https://deno-fetch-h2-repro.dev/";

    let resp = client.get(url).send().await?;
    let t = resp.text().await?;
    println!("✅ {id} {t}");

    Ok(())
}
