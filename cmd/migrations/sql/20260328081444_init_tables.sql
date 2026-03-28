-- Create "companies" table
CREATE TABLE "public"."companies" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "name" text NULL,
  "keycloak_id" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_companies_deleted_at" to table: "companies"
CREATE INDEX "idx_companies_deleted_at" ON "public"."companies" ("deleted_at");
-- Create "users" table
CREATE TABLE "public"."users" (
  "id" uuid NOT NULL,
  "created_at" timestamptz NOT NULL DEFAULT now(),
  "updated_at" timestamptz NOT NULL DEFAULT now(),
  "deleted_at" timestamptz NULL,
  "first_name" text NULL,
  "last_name" text NULL,
  "email" text NULL,
  "keycloak_id" text NULL,
  "stripe_customer_id" text NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_users_deleted_at" to table: "users"
CREATE INDEX "idx_users_deleted_at" ON "public"."users" ("deleted_at");
-- Create "user_companies" table
CREATE TABLE "public"."user_companies" (
  "user_id" uuid NOT NULL,
  "company_id" uuid NOT NULL,
  PRIMARY KEY ("user_id", "company_id"),
  CONSTRAINT "fk_user_companies_company" FOREIGN KEY ("company_id") REFERENCES "public"."companies" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_user_companies_user" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
