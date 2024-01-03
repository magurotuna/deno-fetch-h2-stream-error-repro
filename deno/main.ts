const CONCURRENCY = Math.floor(Deno.env.get("CONCURRENCY") ?? "1");
if (Number.isNaN(CONCURRENCY) || CONCURRENCY <= 0) {
  throw new Error(
    `the provided concurrency is not valid: ${Deno.args.at(0)}`,
  );
}

const client = Deno.createHttpClient({
  http1: true,
  http2: true,
});

const url = "https://deno-fetch-h2-repro.dev";

const promises = [];
for (let i = 0; i < CONCURRENCY; i++) {
  promises.push(request(i));
}

await Promise.allSettled(promises);
console.log("done");

async function request(id: number) {
  const url = "https://deno-fetch-h2-repro.dev";
  const res = await fetch(url, { client }).catch((e) => console.error(`❌ ${id}: ${e}`));
  console.log(`✅ ${id}: ${res.status}`);
}
